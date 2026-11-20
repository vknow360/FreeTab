# FreeTab Design System — MASTER.md
*Generated following the ui-ux-pro-max specification for precision desktop utilities.*

## 1. Product Archetype & Visual Style
- **Archetype**: Precision Hardware & Firmware Engineering Utility (comparable to Balena Etcher, Apple Configurator, Linear, Raycast).
- **Visual Style**: Clean Modern Desktop System Utility. High-contrast typography, 1px crisp borders, subtle elevation shadows, inline technical telemetry HUD, and zero decorative emojis.
- **Target OS**: Windows 10/11 Desktop (Wails / WebView2 Native Shell).

## 2. Color Palette & Tokens
- **Background Main**: `#f8fafc` (Slate 50)
- **Surface Elevated**: `#ffffff` (Pure White)
- **Surface Secondary**: `#f1f5f9` (Slate 100)
- **Border Default**: `#e2e8f0` (Slate 200)
- **Border Subtle**: `#cbd5e1` (Slate 300)
- **Text Primary**: `#0f172a` (Slate 900 - high contrast, authoritative)
- **Text Secondary**: `#475569` (Slate 600 - readable, balanced)
- **Text Muted**: `#64748b` (Slate 500)
- **Accent Primary**: `#0284c7` (Sky 600 - precision engineering blue)
- **Accent Primary Hover**: `#0369a1` (Sky 700)
- **Accent Primary Subtle**: `#e0f2fe` (Sky 100)
- **Success / Connected**: `#059669` (Emerald 600)
- **Warning / Fastboot**: `#d97706` (Amber 600)
- **Destructive / Error**: `#e11d48` (Rose 600)
- **Terminal Dark**: `#090d16` (Obsidian Slate) / Header: `#131b2e`

## 3. Typography
- **Primary Font Stack**: `system-ui, -apple-system, "Segoe UI", Roboto, "Helvetica Neue", sans-serif`
- **Monospace Font Stack**: `"JetBrains Mono", "Cascadia Code", Consolas, "SF Mono", monospace`
- **Hierarchy**:
  - Window Title / Brand: 15px / 700 weight / uppercase tracking 0.05em
  - Section Headings: 20px / 700 weight / letter-spacing -0.02em
  - Subheaders: 14px / 600 weight
  - Body: 13.5px / 400 weight / line-height 1.55
  - Captions / Tech specs: 11.5px / 600 weight monospace

## 4. Anti-Patterns Avoided (Explicit Guardrails)
- ❌ **NO Emojis as primary UI icons**: All icons are pixel-aligned, clean inline SVGs with 1.75px stroke width.
- ❌ **NO Garish colorful card fills**: Backgrounds remain pure white with clean 1px borders and refined micro-badges.
- ❌ **NO Ambiguous button states**: All actions provide spinner feedback, clear disabled states, and explicit confirmation for destructive steps.
- ❌ **NO Boxy, unstyled web forms**: File inputs and directory pickers use native-feeling desktop styling with integrated browse action triggers.
- ❌ **NO Unstructured walls of text**: Instructions are segmented into ordered chronological badges with distinct hardware warning callouts.

## 5. Micro-Interactions & States
- Interactive buttons have active `transform: translateY(1px)` tactile depress response.
- Connected timeline stepper connects steps with a continuous progress track.
- Telemetry badges pulse subtly when connection status changes or hardware handshake completes.
- Console logs feature monospace formatting, syntax color coding, timestamps, and smooth auto-scrolling.
