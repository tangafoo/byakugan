// Command corpus is byakugan's brain-side CLI — the terminal window into the
// law before any app exists. Phase A scope: load a statute file, validate it,
// print what's inside. Later phases bolt on `embed`, `query`, `eval`.
//
// Run it: go run ./cmd/corpus load data/raw/rta1987.sample.jsonl
//
//	go run ./cmd/corpus load --lang en data/raw/rta1987.sample.jsonl
package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"
	"text/tabwriter"

	"github.com/joho/godotenv"

	"byakugan/internal/corpus"
	"byakugan/internal/eval"
	"byakugan/internal/store"
	"byakugan/internal/voyage"
)

// loadVoyageKey best-effort loads server/.env, then reads the key from the
// environment. godotenv.Load looks for a `.env` in the current working dir
// (run the CLI from server/). We ignore its error on purpose: in prod there's
// no .env file — the platform sets VOYAGE_API_KEY directly — so a missing file
// isn't fatal; a missing *key* is.
func loadVoyageKey() (string, error) {
	_ = godotenv.Load()
	key := os.Getenv("VOYAGE_API_KEY")
	if key == "" {
		return "", fmt.Errorf("VOYAGE_API_KEY not set")
	}
	return key, nil
}

// dsn is the database connection string. Env override first (so prod/Railway can
// inject its own), local docker default otherwise. Note port 5433 — we remapped
// off 5432 to dodge the Supabase containers. A DSN packs host, port, user, pass,
// and db into one URL: postgres://user:pass@host:port/dbname.
func dsn() string {
	if v := os.Getenv("DATABASE_URL"); v != "" {
		return v
	}
	return "postgres://byakugan:byakugan@localhost:5433/byakugan"
}

func main() {
	// os.Args[0] is the program name; [1] is the subcommand. Go's stdlib has no
	// built-in subcommand router (no cobra here yet — staying dependency-free),
	// so we dispatch by hand. Honest and tiny.
	if len(os.Args) < 2 {
		usage(os.Stderr)
		os.Exit(2)
	}

	switch os.Args[1] {
	case "load":
		if err := runLoad(os.Args[2:]); err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}
	case "migrate":
		if err := runMigrate(); err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}
	case "embed":
		if err := runEmbed(os.Args[2:]); err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}
	case "query":
		if err := runQuery(os.Args[2:]); err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}
	case "eval":
		if err := runEval(os.Args[2:]); err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}
	case "--help", "-h", "help":
		usage(os.Stdout)
		os.Exit(0)
	default:
		usage(os.Stderr)
		os.Exit(2)
	}
}

func usage(w io.Writer) {
	fmt.Fprint(w, `
	byakugan! - (▰˘◡˘▰) made by Gafu
	-------------------------------------------------------------
	how to use:	./cmd/corpus	load | migrate | embed	[filename]
	
	load	[--lang ms|en]	[filename.jsonl]	inspect a corpus file
	migrate						create the pgvector schema
	embed	[filename.jsonl]			embed a corpus file via Voyage
	query	[--limit default 5]	[question]	which law applies to given question
	eval	[--k limit default 5]	[filename.eval.jsonl]	check search passes eval gate
	`)
}

// runEmbed — YOUR turn. See the spec in chat. The pieces you have to work with:
//   - corpus.LoadFile(path) ([]corpus.Chunk, error)   — load + validate
//   - loadVoyageKey() (string, error)                 — the key (plumbing done)
//   - voyage.New(key) *voyage.Client                  — make the client once
//   - client.Embed(ctx, voyage.Document, texts) ([][]float32, error)
//
// Goal for now: print how many vectors came back and the length of the first
// one. (Saving them to Postgres is the next step, 2c.)
func runEmbed(args []string) error {
	ctx := context.Background()
	fs := flag.NewFlagSet("embed", flag.ExitOnError)
	if err := fs.Parse(args); err != nil {
		return err
	}

	if fs.NArg() != 1 {
		return fmt.Errorf("needs exactly 1 filename as argument. What do you want to embed ?")
	}

	chunks, err := corpus.LoadFile(fs.Arg(0))
	if err != nil {
		return fmt.Errorf("runEmbed: %w", err)
	}

	key, err := loadVoyageKey()
	if err != nil {
		return fmt.Errorf("⊙︿⊙ trouble loading voyage key: %w", err)
	}

	voyageClient := voyage.New(key)

	var texts []string
	for _, c := range chunks {
		texts = append(texts, c.Text)
	}

	vectors, err := voyageClient.Embed(ctx, voyage.Document, texts)
	if err != nil {
		return fmt.Errorf("runEmbed: %w", err)
	}

	if len(vectors) == 0 {
		return fmt.Errorf("nothing got embedded ʘ︵ʘ")
	}
	fmt.Printf("Received %d vectors, each of length %d\n", len(vectors), len(vectors[0]))

	st, err := store.Connect(ctx, dsn())
	if err != nil {
		return fmt.Errorf("could not connect store/DB in runEmbed: %w", err)
	}
	defer st.Close()

	// for error tracking only - hopefully all succeeds !
	var failedChunkUploads []string
	for i, c := range chunks {
		if err := st.UpsertChunk(ctx, c, vectors[i]); err != nil {
			fmt.Fprintf(os.Stderr, "[warn] failed to upsert chunk: %v\n", err)
			failedChunkUploads = append(failedChunkUploads, fmt.Sprintf("failed to upsert chunk %s: %s", c.ID, c.Heading))
			continue
		}
	}

	if len(failedChunkUploads) == len(chunks) {
		fmt.Printf("failed to upload any chunks ʘ︵ʘ")
		os.Exit(1)
	}

	if len(failedChunkUploads) > 0 {
		fmt.Printf("embed: done - with %d chunks failed to upload\n", len(failedChunkUploads))
		for _, msg := range failedChunkUploads {
			fmt.Printf("[%s]\n", msg)
		}
	} else {
		fmt.Printf("embed: all chunks uploaded successfully (◍•ᴗ•◍)❤\n")
	}
	return nil
}

// runMigrate opens the DB and applies the schema. context.Background() is the
// root context — "no deadline, no cancellation"; fine for a one-shot CLI command.
// A long-lived server would derive per-request contexts from it instead.
func runMigrate() error {
	ctx := context.Background()
	st, err := store.Connect(ctx, dsn())
	if err != nil {
		return err
	}
	defer st.Close()
	if err := st.Migrate(ctx); err != nil {
		return err
	}
	fmt.Println("migrated: pgvector enabled, chunks table ready")
	return nil
}

// runLoad parses this subcommand's own flags. flag.NewFlagSet gives each
// subcommand its own flag namespace — the idiom for `git commit -m` style CLIs
// where each verb has different options.
func runLoad(args []string) error {
	fs := flag.NewFlagSet("load", flag.ExitOnError)
	langStr := fs.String("lang", "", "filter to one language: ms | en (empty = all)")

	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 1 {
		return fmt.Errorf("expected exactly one file argument")
	}

	// The language selector you asked for, in its first form: scope the view to
	// one language. Later this same Lang threads into the Ask request to pick the
	// language of the *framed answer*.
	var want corpus.Lang
	if *langStr != "" {
		want = corpus.Lang(*langStr)
		if !want.Valid() {
			return fmt.Errorf("invalid --lang %q (please choose either ms or en ಥ‿ಥ)", *langStr)
		}
	}

	chunks, err := corpus.LoadFile(fs.Arg(0))
	if err != nil {
		return err
	}

	// tabwriter aligns columns by padding to the widest cell — like piping
	// through `column -t`, but built into the stdlib. Write tab-separated, it
	// handles the spacing on Flush.
	tw := tabwriter.NewWriter(os.Stdout, 0, 2, 2, ' ', 0)
	fmt.Fprintln(tw, "SECTION\tAUTH\tSTATE\tLANG\tVERIFIED\tHEADING")

	var shown, verified int
	for _, c := range chunks {
		if want != "" && c.Lang != want {
			continue
		}
		shown++
		if c.Verified {
			verified++
		}
		mark := "PENDING"
		if c.Verified {
			mark = "ok"
		}
		fmt.Fprintf(tw, "s%s\t%s\t%s\t%s\t%s\t%s\n",
			c.Section, c.Authority, c.State, c.Lang, mark, c.Heading)
	}
	tw.Flush()

	fmt.Printf("\n%d chunks shown (%d verified, %d pending verbatim sourcing)\n",
		shown, verified, shown-verified)
	return nil
}

func runQuery(args []string) error {
	ctx := context.Background()
	fs := flag.NewFlagSet("query", flag.ExitOnError)
	limitStr := fs.Int("limit", 5, "number of results to return")
	langStr := fs.String("lang", "en", "choose language [en | ms] (en is default)")

	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("[query]: %w", err)
	}

	if fs.NArg() < 1 {
		return fmt.Errorf("What question do you want to ask byakugan ? ᕕ( ಠ‿ಠ)ᕗ\nPlease specify a question as an argument")
	}

	var want corpus.Lang
	if *langStr != "" {
		want = corpus.Lang(*langStr)
		if !want.Valid() {
			return fmt.Errorf("invalid --lang %q, sorry! we only support en | ms. For the time being.", *langStr)
		}
	} else {
		want = "en"
	}

	question := strings.Join(fs.Args(), " ")

	key, err := loadVoyageKey()
	if err != nil {
		return fmt.Errorf("⊙︿⊙ trouble loading voyage key: %w", err)
	}

	voyageClient := voyage.New(key)
	vectors, err := voyageClient.Embed(ctx, voyage.Query, []string{question})
	if err != nil {
		return err
	}

	if len(vectors) == 0 {
		return fmt.Errorf("nothing got embedded ʘ︵ʘ")
	}

	st, err := store.Connect(ctx, dsn())
	if err != nil {
		return fmt.Errorf("could not connect store/DB in runQuery: %w", err)
	}
	defer st.Close()

	results, err := st.Search(ctx, vectors[0], *limitStr, want)
	if err != nil {
		return fmt.Errorf("[query] %w", err)
	}

	if len(results) == 0 {
		return fmt.Errorf("No results for question.\nFlags set ->  LIMIT: %d\tLANG: %q", *limitStr, *langStr)
	}

	tw := tabwriter.NewWriter(os.Stdout, 0, 3, 3, ' ', 0)
	fmt.Fprintln(tw, "ID\tSECTION\tHEADING\tLANG\tDISTANCE\tTEXT")

	for _, h := range results {
		runes := []rune(h.Text)
		if len(runes) > 80 {
			runes = runes[:80]
		}

		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%.5f\t%s\n", h.ID, h.Section, h.Heading, h.Lang, h.Distance, string(runes))
	}
	tw.Flush()

	fmt.Printf("\n%d found. Question was %q?\n", len(results), question)

	return nil
}

func runEval(args []string) error {
	ctx := context.Background()

	fs := flag.NewFlagSet("eval", flag.ExitOnError)
	k := fs.Int("k", 5, "top-k to search")
	rerank := fs.Bool("rerank", false, "rerank with voyage before scoring")

	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("Parsing error: %w", err)
	}
	if fs.NArg() != 1 {
		return fmt.Errorf("Needs an eval file path")
	}

	cases, err := eval.LoadFile(fs.Arg(0))
	if err != nil {
		return fmt.Errorf("troubling reading file and converting to cases: %w", err)
	}

	key, err := loadVoyageKey()
	if err != nil {
		return fmt.Errorf("could not load voyage API key: %w", err)
	}

	voyageClient := voyage.New(key)

	var questions []string
	for _, c := range cases {
		questions = append(questions, c.Question)
	}

	vectors, err := voyageClient.Embed(ctx, voyage.Query, questions)
	if err != nil {
		return err
	}

	st, err := store.Connect(ctx, dsn())
	if err != nil {
		return fmt.Errorf("could not correct to store/DB: %w", err)
	}
	defer st.Close()

	var passedCases int
	var allSearchScores []float32
	var validCases float32

	for i, tc := range cases {
		fmt.Println("\n------------")
		fmt.Printf("CASE [%q]\n", tc.ID)
		searchK := *k

		if *rerank {
			searchK = 20
		}

		hits, err := st.Search(ctx, vectors[i], searchK, tc.Lang)
		if err != nil {
			return fmt.Errorf("trouble searching: %w", err)
		}

		if *rerank {
			var hitTexts []string
			for _, h := range hits {
				hitTexts = append(hitTexts, h.Text)
			}

			reranked, err := voyageClient.Rerank(ctx, tc.Question, hitTexts, *k)
			if err != nil {
				return fmt.Errorf("Could not run rerank: %w", err)
			}

			newHitsSlice := make([]store.Hit, 0, len(reranked))

			for _, rr := range reranked {
				newHitsSlice = append(newHitsSlice, hits[rr.Index])
			}
			hits = newHitsSlice
		}

		fmt.Printf("%d hits found for question %q\n", len(hits), tc.Question)

		fmt.Printf("LANG: %q\n", tc.Lang)
		fmt.Println(">>>")

		var matchedHits int
		var searchPrecision float32

		for j, h := range hits {
			position := j + 1
			fmt.Fprintf(os.Stdout, "found %s - DISTANCE: [%.4f] ", h.Section, h.Distance)

			if slices.Contains(tc.ExpectSections, h.Section) {
				fmt.Print("✅ was expected. PASS")
				matchedHits++

				searchPrecision += float32(matchedHits) / float32(position)
			} else {
				fmt.Print("❌ not expected. FAIL")
			}

			fmt.Printf(" ✯ RANKED %d", position)
			fmt.Println()
		}

		if matchedHits > 0 {
			searchScore := searchPrecision / float32(len(tc.ExpectSections))
			allSearchScores = append(allSearchScores, searchScore)

			fmt.Printf("\nSCORE OF RESULTS: %.2f\n\n", searchScore)
		} else {
			fmt.Printf("\nNO RESULTS\n\n")
		}

		fmt.Println(">>>")

		switch {
		case tc.ShouldFind:
			validCases++
			fmt.Printf("Found %d/%d %v EXPECTED SECTIONS (◍•ᴗ•◍)\n", matchedHits, len(tc.ExpectSections), tc.ExpectSections)

			if matchedHits == len(tc.ExpectSections) {
				fmt.Fprintf(os.Stdout, "⭐️ [PASS] expected to find %d, actually found %d.", len(tc.ExpectSections), matchedHits)
				passedCases++
			} else {
				fmt.Fprintf(os.Stdout, "[FAIL] Did not find all expected sections")
			}
		case !tc.ShouldFind && len(hits) > 0:
			fmt.Fprintln(os.Stdout, "[FAIL] Did not expect to find any section actually, but search returned some. ಠ_ಠ")
		case !tc.ShouldFind && matchedHits == 0:
			passedCases++
			fmt.Fprintln(os.Stdout, "[PASS] did not expect to find any. Truly did not find any.")
		}
	}

	var searchQuality float32
	for _, score := range allSearchScores {
		searchQuality += score
	}

	searchQuality = searchQuality / validCases

	fmt.Println("============")
	fmt.Printf("\nSummary: %d/%d passed\n\n", passedCases, len(cases))
	fmt.Printf("ALL SEARCH SCORES: %v\n", allSearchScores)
	fmt.Printf("Mean Average Precision - A.K.A how well Byakugan is at search ᕙ(◕ل͜◕)ᕗ = %.4f\n", searchQuality)
	return nil
}
