// Command corpus is byakugan's brain-side CLI — the terminal window into the
// law before any app exists. Phase A scope: load a statute file, validate it,
// print what's inside. Later phases bolt on `embed`, `query`, `eval`.
//
// Run it: go run ./cmd/corpus load data/raw/rta1987.sample.jsonl
//         go run ./cmd/corpus load --lang en data/raw/rta1987.sample.jsonl
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"text/tabwriter"

	"byakugan/internal/corpus"
	"byakugan/internal/store"
)

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
		usage()
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
	default:
		usage()
		os.Exit(2)
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: corpus <load|migrate> ...")
	fmt.Fprintln(os.Stderr, "  load [--lang ms|en] <file.jsonl>   inspect a corpus file")
	fmt.Fprintln(os.Stderr, "  migrate                            create the pgvector schema")
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
	langStr := fs.String("lang", "", "filter to one language: ms|en (empty = all)")
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
			return fmt.Errorf("invalid --lang %q (want ms or en)", *langStr)
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
