# OpenAid.se — Visual Style Guide

> Reverse-engineered from [openaid.se](https://openaid.se/en/countries-and-regions?years=2025&tab=strategies), March 2026.
> OpenAid is Sweden's public aid transparency tracker, built by Sida (Swedish International Development Cooperation Agency).

---

## 1. Design Philosophy

OpenAid follows a **Scandinavian data-transparency** aesthetic: clean, restrained, and information-first. The design avoids decoration in favour of generous whitespace, thin dividers, and a near-monochrome palette that lets the data breathe. The overall feel is closer to a government research tool than a marketing site — deliberate, trustworthy, and highly legible.

---

## 2. Colour Palette

### Primary Palette

| Role | Hex | Usage |
|------|-----|-------|
| **Background** | `#FFFFFF` | Page & card backgrounds |
| **Surface / Card** | `#FFFFFF` | Content cards, panels |
| **Text — Primary** | `#1A1A1A` | Headings, labels, axis values |
| **Text — Secondary** | `#666666` | Subtext, helper text, "Choose view" |
| **Text — Tertiary** | `#999999` | Info bar captions, footnotes |

### Chart Palette

| Role | Hex | Usage |
|------|-----|-------|
| **Line Stroke** | `#1B3A2A` | Dark forest green — area chart line |
| **Area Fill** | `#D6ECDA` | Light mint green — area under curve |
| **Highlight Band** | `#E8E8E8` / 50% opacity | Grey overlay for selected year column |
| **Data Point (active)** | `#3B3BF9` | Indigo/blue dot — hovered or selected data point |
| **Data Point Ring** | `#FFFFFF` | White ring around active dot |

### UI Accents

| Role | Hex | Usage |
|------|-----|-------|
| **Divider** | `#E5E5E5` | Thin 1px horizontal rules between sections |
| **Border — Active Control** | `#1A1A1A` | Selected pill / toggle border |
| **Border — Inactive Control** | `#E0E0E0` | Unselected pill border |
| **Card Border** | `#EBEBEB` | Subtle card outlines, tooltip card |

---

## 3. Typography

OpenAid uses a clean sans-serif stack. The typeface appears to be **system sans-serif** or a geometric grotesque similar to Inter, Söhne, or National.

### Font Stack (Recommended)

```css
font-family: 'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
```

### Type Scale

| Element | Size (approx) | Weight | Colour | Notes |
|---------|--------------|--------|--------|-------|
| **Page Heading** (H1) | 24–28px | `600` Semi-bold | `#1A1A1A` | e.g. "Development of aid" |
| **Section Label** | 13–14px | `400` Regular | `#666666` | e.g. "Choose view" |
| **Pill / Toggle Text** | 14px | `500` Medium | `#1A1A1A` | e.g. "Line chart", "Table" |
| **Tooltip Title** | 15–16px | `600` Semi-bold | `#1A1A1A` | e.g. "Year 2025" |
| **Tooltip Value** | 13–14px | `400` Regular | `#666666` | e.g. "40.42 bn SEK" |
| **Axis Labels** | 12–13px | `400` Regular | `#666666` | e.g. "50 bn SEK", "2014" |
| **Footnote / Info** | 13px | `400` Regular | `#999999` | e.g. "The graph about yearly data" |

### Key Typographic Choices

- No uppercase / all-caps in headings — sentence case throughout
- No decorative fonts — purely functional
- Numbers are **tabular** (mono-spaced digits) on axis labels for clean vertical alignment
- Currency abbreviations inline: "bn SEK", "M SEK"
- Minimal letterspacing — default or slightly tighter

---

## 4. Spacing & Layout

### Grid & Container

- **Max content width**: ~1100–1200px, centered
- **Card padding**: `24–32px` all sides
- **Section spacing**: `16–24px` between stacked sections

### Whitespace Principles

- Charts have generous top/bottom padding (~40–60px within the card)
- Y-axis labels sit outside the chart area with ~12px clearance
- X-axis labels have ~16px clearance below the chart baseline
- Tooltip card has ~16–20px internal padding

---

## 5. Borders & Dividers

| Element | Style |
|---------|-------|
| **Horizontal section divider** | `1px solid #E5E5E5` — full width |
| **Card border** | `1px solid #EBEBEB` with `border-radius: 12–16px` |
| **Chart gridlines** | `1px dashed #E5E5E5` or very faint solid — horizontal only |
| **Pill / toggle border (active)** | `1.5px solid #1A1A1A`, `border-radius: 999px` (fully rounded) |
| **Pill / toggle border (inactive)** | `1.5px solid #E0E0E0`, same radius |

### Border Radius

| Element | Radius |
|---------|--------|
| Main content card | `12–16px` |
| Tooltip popup | `8–12px` |
| Pills / toggles | `999px` (full pill shape) |
| Buttons | `8px` or `999px` depending on context |

---

## 6. Components

### 6a. Toggle Pill Group

The "Choose view" selector (Line chart | Table) uses a pill-based toggle:

```
┌──────────────┐  ┌──────────┐
│ ✓ Line chart │  │  Table   │
└──────────────┘  └──────────┘
  ↑ active           ↑ inactive
  border: 1.5px      border: 1.5px
  #1A1A1A             #E0E0E0
  bg: #FFFFFF         bg: #F8F8F8
```

- Each pill includes an **icon** (chart icon, grid icon) to the left of the label
- Icons are 16–18px, matching text colour
- Gap between icon and text: `6–8px`
- Gap between pills: `8px`
- Pill padding: `8px 16px`

### 6b. Tooltip / Info Card

The "Year 2025 — 40.42 bn SEK" popup:

- Background: `#FFFFFF`
- Border: `1px solid #EBEBEB`
- Border-radius: `8–12px`
- Box-shadow: `0 2px 8px rgba(0,0,0,0.06)`
- Padding: `16–20px`
- Title: Semi-bold, primary colour
- Value: Regular weight, secondary colour
- Positioned above or near the hovered data point

### 6c. Footnote Bar

"ⓘ The graph about yearly data" at the bottom:

- Full-width, below chart
- Top border: `1px solid #E5E5E5`
- Icon: info circle `ⓘ`, 16px, colour `#999999`
- Text: 13px, regular, `#999999`
- Padding: `16px 0`

---

## 7. Chart / Data Visualisation

### Area Chart (the "Development of Aid" graph)

This is the centrepiece component. Key characteristics:

**Line**
- Stroke colour: `#1B3A2A` (dark forest green)
- Stroke width: `2–2.5px`
- Smooth curve interpolation (monotone or catmull-rom, not linear)
- No data point markers visible by default

**Fill**
- Gradient or solid fill from line to x-axis baseline
- Colour: `#D6ECDA` (light mint green), ~30–40% opacity
- Creates a soft area chart effect

**Active Data Point**
- Appears on hover/selection
- Outer ring: `#FFFFFF`, 4px
- Inner dot: `#3B3BF9` (indigo/blue), 8px diameter
- Total diameter including ring: ~16px

**Selected Year Highlight**
- Vertical band/column behind the selected year
- Colour: `#E8E8E8` at ~50% opacity
- Spans full chart height
- Width: matches one year interval on x-axis

**Axes**
- Y-axis: labelled in `bn SEK` increments (0, 10, 20, 30, 40, 50)
- X-axis: yearly labels (1998–2026)
- Axis labels: 12–13px, `#666666`
- Selected year (2025) on x-axis: **bold weight**, `#1A1A1A`
- No visible axis lines — labels float with subtle gridlines only
- Gridlines: horizontal only, `1px`, `#F0F0F0`, dashed or very light solid

---

## 8. Shadows & Elevation

The design is extremely flat. Shadows are minimal:

| Element | Shadow |
|---------|--------|
| Content cards | `0 1px 3px rgba(0,0,0,0.04)` or none |
| Tooltip popup | `0 2px 8px rgba(0,0,0,0.06)` |
| Hover states | `0 2px 6px rgba(0,0,0,0.08)` |

---

## 9. Iconography

- Style: **outline/line icons**, 1.5–2px stroke
- Size: 16–18px in UI controls
- Colour: matches adjacent text colour
- Chart icon (📈): line-style, used in "Line chart" pill
- Table icon (⊞): grid-style, used in "Table" pill
- Info icon (ⓘ): circle with "i", used in footnotes

---

## 10. CSS Custom Properties (Design Tokens)

```css
:root {
  /* Colours */
  --color-bg:              #FFFFFF;
  --color-text-primary:    #1A1A1A;
  --color-text-secondary:  #666666;
  --color-text-tertiary:   #999999;
  --color-border:          #E5E5E5;
  --color-border-light:    #EBEBEB;
  --color-border-active:   #1A1A1A;

  /* Chart */
  --chart-line:            #1B3A2A;
  --chart-area-fill:       #D6ECDA;
  --chart-dot-active:      #3B3BF9;
  --chart-highlight-band:  rgba(232, 232, 232, 0.5);
  --chart-gridline:        #F0F0F0;

  /* Typography */
  --font-family:           'Inter', -apple-system, BlinkMacSystemFont,
                           'Segoe UI', Roboto, sans-serif;
  --font-size-h1:          1.625rem;   /* ~26px */
  --font-size-body:        0.875rem;   /* 14px */
  --font-size-small:       0.8125rem;  /* 13px */
  --font-size-axis:        0.75rem;    /* 12px */
  --font-weight-regular:   400;
  --font-weight-medium:    500;
  --font-weight-semibold:  600;

  /* Spacing */
  --space-xs:              4px;
  --space-sm:              8px;
  --space-md:              16px;
  --space-lg:              24px;
  --space-xl:              32px;
  --space-2xl:             48px;

  /* Borders */
  --radius-sm:             8px;
  --radius-md:             12px;
  --radius-lg:             16px;
  --radius-pill:           999px;
  --border-width:          1px;
  --border-width-active:   1.5px;

  /* Shadows */
  --shadow-card:           0 1px 3px rgba(0, 0, 0, 0.04);
  --shadow-tooltip:        0 2px 8px rgba(0, 0, 0, 0.06);
  --shadow-hover:          0 2px 6px rgba(0, 0, 0, 0.08);
}
```

---

## 11. Interaction Patterns

| Interaction | Behaviour |
|-------------|-----------|
| **Chart hover** | Data point dot appears; tooltip card fades in; year column highlights |
| **Toggle pill** | Instant switch, no transition animation; border colour changes |
| **Year selection** | Click on x-axis label or chart area to select year |
| **Responsive** | Chart scales proportionally; x-axis labels may thin out on mobile |

---

## 12. Guiding Principles Summary

1. **Data first** — every visual choice serves readability and comprehension
2. **Restrained palette** — near-monochrome with one accent green for the data layer
3. **Generous whitespace** — never cramped, always breathing room
4. **Thin dividers over boxes** — hairline rules separate sections, not heavy borders
5. **Minimal shadows** — almost perfectly flat, elevation only where absolutely needed
6. **Functional typography** — clean sans-serif, clear hierarchy, no flourish
7. **Trustworthy tone** — appropriate for a government transparency platform

---

*Source: Visual analysis of openaid.se, March 2026. Exact values are best-estimate approximations from screenshot inspection. For pixel-perfect implementation, inspect the live site's CSS via browser developer tools.*
