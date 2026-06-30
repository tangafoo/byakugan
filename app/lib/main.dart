import 'dart:ui';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';

void main() {
  runApp(const ByakuganApp());
}

/// ── Palette ──────────────────────────────────────────────────────────────
/// Cool, dark, blue. One hot accent (protest red) for tension.
/// Chrome is loud; the law (later) will be calm. This is the loud surface.
class Ink {
  static const bg0 = Color(0xFF060912); // near-black navy
  static const bg1 = Color(0xFF0A1326); // deep blue
  static const glow = Color(0xFF2E6BFF); // electric blue (search glow)
  static const glass = Color(0x14FFFFFF); // 8% white — glass fill
  static const stroke = Color(0x1FFFFFFF); // 12% white — glass edge
  static const text = Color(0xFFEAF0FF);
  static const sub = Color(0x99EAF0FF);

  // Per-category accents — graffiti energy, each its own colour.
  static const road = Color(0xFF35C4FF); // cyan
  static const detained = Color(0xFFFF3B6B); // protest red
  static const door = Color(0xFFB983FF); // violet
  static const stand = Color(0xFFFFC24B); // amber
}

class ByakuganApp extends StatelessWidget {
  const ByakuganApp({super.key});

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'byakugan',
      debugShowCheckedModeBanner: false,
      theme: ThemeData(
        brightness: Brightness.dark,
        scaffoldBackgroundColor: Ink.bg0,
        fontFamily: 'Roboto',
        useMaterial3: true,
      ),
      home: const HomeScreen(),
    );
  }
}

/// A one-tap scenario category. Routes "situation first" — tapping it will
/// later ask "who's in front of you?" (PDRM / JPJ / DBKL / religious).
class Category {
  final String en;
  final String bm;
  final IconData icon;
  final Color accent;
  const Category(this.en, this.bm, this.icon, this.accent);
}

const _categories = <Category>[
  Category('Road stop', 'Disekat di jalan', Icons.directions_car_filled, Ink.road),
  Category('Detained?', 'Ditahan?', Icons.pan_tool, Ink.detained),
  Category('At my door', 'Di pintu rumah', Icons.door_front_door, Ink.door),
  Category('Where I stand', 'Di mana saya berdiri', Icons.local_fire_department, Ink.stand),
];

class HomeScreen extends StatefulWidget {
  const HomeScreen({super.key});

  @override
  State<HomeScreen> createState() => _HomeScreenState();
}

class _HomeScreenState extends State<HomeScreen> {
  bool _bm = false; // false = EN, true = BM. App will remember this later.

  String get _placeholder => _bm
      ? 'apa nak buat bila polis sekat kereta?'
      : 'what to do when cops stop you on the road?';

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: Stack(
        children: [
          // Cool blue base gradient.
          const _Backdrop(),
          // Soft electric glow rising from the search bar.
          const Positioned(
            left: 0,
            right: 0,
            bottom: -120,
            child: _Glow(),
          ),
          SafeArea(
            child: Padding(
              padding: const EdgeInsets.fromLTRB(20, 12, 20, 20),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.stretch,
                children: [
                  _TopBar(bm: _bm, onToggle: (v) => setState(() => _bm = v)),
                  const SizedBox(height: 18),
                  // Tiles own the whole upper canvas now (no headline).
                  Expanded(
                    child: GridView.count(
                      crossAxisCount: 2,
                      mainAxisSpacing: 14,
                      crossAxisSpacing: 14,
                      childAspectRatio: 0.92,
                      physics: const NeverScrollableScrollPhysics(),
                      children: [
                        for (final c in _categories)
                          _CategoryTile(category: c, bm: _bm),
                      ],
                    ),
                  ),
                  const SizedBox(height: 14),
                  _SearchBar(hint: _placeholder),
                ],
              ),
            ),
          ),
        ],
      ),
    );
  }
}

class _Backdrop extends StatelessWidget {
  const _Backdrop();
  @override
  Widget build(BuildContext context) {
    return Container(
      decoration: const BoxDecoration(
        gradient: LinearGradient(
          begin: Alignment.topCenter,
          end: Alignment.bottomCenter,
          colors: [Ink.bg1, Ink.bg0],
        ),
      ),
    );
  }
}

class _Glow extends StatelessWidget {
  const _Glow();
  @override
  Widget build(BuildContext context) {
    return IgnorePointer(
      child: Container(
        height: 360,
        decoration: const BoxDecoration(
          gradient: RadialGradient(
            radius: 0.7,
            colors: [Color(0x552E6BFF), Color(0x00060912)],
          ),
        ),
      ),
    );
  }
}

class _TopBar extends StatelessWidget {
  final bool bm;
  final ValueChanged<bool> onToggle;
  const _TopBar({required this.bm, required this.onToggle});

  @override
  Widget build(BuildContext context) {
    return Row(
      children: [
        // Wordmark — lowercase, tight, a little loud.
        const Text(
          'byakugan',
          style: TextStyle(
            color: Ink.text,
            fontSize: 22,
            fontWeight: FontWeight.w800,
            letterSpacing: -0.5,
          ),
        ),
        const Spacer(),
        _LangToggle(bm: bm, onToggle: onToggle),
        const SizedBox(width: 12),
        // Profile stub (no auth yet — mocked).
        Container(
          height: 38,
          width: 38,
          decoration: BoxDecoration(
            color: Ink.glass,
            shape: BoxShape.circle,
            border: Border.all(color: Ink.stroke),
          ),
          child: const Icon(Icons.person_outline, size: 20, color: Ink.sub),
        ),
      ],
    );
  }
}

class _LangToggle extends StatelessWidget {
  final bool bm;
  final ValueChanged<bool> onToggle;
  const _LangToggle({required this.bm, required this.onToggle});

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 38,
      padding: const EdgeInsets.all(3),
      decoration: BoxDecoration(
        color: Ink.glass,
        borderRadius: BorderRadius.circular(20),
        border: Border.all(color: Ink.stroke),
      ),
      child: Row(
        children: [
          _segment('BM', bm, () => onToggle(true)),
          _segment('EN', !bm, () => onToggle(false)),
        ],
      ),
    );
  }

  Widget _segment(String label, bool active, VoidCallback onTap) {
    return GestureDetector(
      onTap: () {
        HapticFeedback.selectionClick();
        onTap();
      },
      child: AnimatedContainer(
        duration: const Duration(milliseconds: 160),
        padding: const EdgeInsets.symmetric(horizontal: 14),
        alignment: Alignment.center,
        decoration: BoxDecoration(
          color: active ? Ink.glow : Colors.transparent,
          borderRadius: BorderRadius.circular(16),
        ),
        child: Text(
          label,
          style: TextStyle(
            color: active ? Colors.white : Ink.sub,
            fontWeight: FontWeight.w700,
            fontSize: 13,
          ),
        ),
      ),
    );
  }
}

class _CategoryTile extends StatelessWidget {
  final Category category;
  final bool bm;
  const _CategoryTile({required this.category, required this.bm});

  @override
  Widget build(BuildContext context) {
    return ClipRRect(
      borderRadius: BorderRadius.circular(22),
      child: BackdropFilter(
        filter: ImageFilter.blur(sigmaX: 12, sigmaY: 12),
        child: Material(
          color: Colors.transparent,
          child: InkWell(
            onTap: () {
              HapticFeedback.lightImpact();
              ScaffoldMessenger.of(context)
                ..hideCurrentSnackBar()
                ..showSnackBar(SnackBar(
                  behavior: SnackBarBehavior.floating,
                  backgroundColor: Ink.bg1,
                  content: Text(
                    'Next: "who is in front of you?" → ${category.en}',
                    style: const TextStyle(color: Ink.text),
                  ),
                ));
            },
            child: Container(
              padding: const EdgeInsets.all(18),
              decoration: BoxDecoration(
                gradient: LinearGradient(
                  begin: Alignment.topLeft,
                  end: Alignment.bottomRight,
                  colors: [
                    category.accent.withValues(alpha: 0.18),
                    Ink.glass,
                  ],
                ),
                borderRadius: BorderRadius.circular(22),
                border: Border.all(color: Ink.stroke),
              ),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Container(
                    height: 46,
                    width: 46,
                    decoration: BoxDecoration(
                      color: category.accent.withValues(alpha: 0.16),
                      borderRadius: BorderRadius.circular(14),
                    ),
                    child: Icon(category.icon, color: category.accent, size: 24),
                  ),
                  const Spacer(),
                  Text(
                    bm ? category.bm : category.en,
                    style: const TextStyle(
                      color: Ink.text,
                      fontSize: 19,
                      fontWeight: FontWeight.w800,
                      letterSpacing: -0.3,
                      height: 1.05,
                    ),
                  ),
                  const SizedBox(height: 4),
                  Container(height: 3, width: 26, color: category.accent),
                ],
              ),
            ),
          ),
        ),
      ),
    );
  }
}

class _SearchBar extends StatelessWidget {
  final String hint;
  const _SearchBar({required this.hint});

  @override
  Widget build(BuildContext context) {
    return Container(
      decoration: BoxDecoration(
        borderRadius: BorderRadius.circular(26),
        boxShadow: [
          BoxShadow(
            color: Ink.glow.withValues(alpha: 0.35),
            blurRadius: 32,
            spreadRadius: -6,
          ),
        ],
      ),
      child: ClipRRect(
        borderRadius: BorderRadius.circular(26),
        child: BackdropFilter(
          filter: ImageFilter.blur(sigmaX: 18, sigmaY: 18),
          child: Container(
            padding: const EdgeInsets.fromLTRB(18, 6, 6, 6),
            decoration: BoxDecoration(
              color: const Color(0x1FFFFFFF),
              borderRadius: BorderRadius.circular(26),
              border: Border.all(color: const Color(0x33FFFFFF)),
            ),
            child: Row(
              children: [
                const Icon(Icons.search, color: Ink.sub, size: 22),
                const SizedBox(width: 12),
                Expanded(
                  child: TextField(
                    style: const TextStyle(color: Ink.text, fontSize: 15),
                    cursorColor: Ink.glow,
                    decoration: InputDecoration(
                      isCollapsed: true,
                      border: InputBorder.none,
                      hintText: hint,
                      hintStyle: const TextStyle(color: Ink.sub, fontSize: 15),
                    ),
                  ),
                ),
                const SizedBox(width: 8),
                Container(
                  height: 44,
                  width: 44,
                  decoration: const BoxDecoration(
                    color: Ink.glow,
                    shape: BoxShape.circle,
                  ),
                  child: const Icon(Icons.arrow_upward, color: Colors.white, size: 22),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }
}
