export const MODEL_COLORS: Record<string, string> = {
  "claude-sonnet-4": "#c4928a",
  "claude-sonnet-4-6": "#c4928a",
  "claude-opus-4": "#3b3bf9",
  "claude-opus-4-6": "#3b3bf9",
  "claude-haiku-4-5": "#8bb8a8",
  "claude-haiku-4-5-20251001": "#8bb8a8",
  "gpt-4o": "#d4c5a9",
  "gpt-4o-mini": "#d4c5a9",
};

export function getModelColor(model: string): string {
  if (MODEL_COLORS[model]) return MODEL_COLORS[model];
  // Try prefix match
  for (const [key, color] of Object.entries(MODEL_COLORS)) {
    if (model.startsWith(key)) return color;
  }
  return "#b0b0b0";
}

export function getModelShortName(model: string): string {
  // "claude-sonnet-4-20250514" → "Claude Sonnet 4"
  const m = model
    .replace(/-\d{8}$/, "") // strip date suffix
    .replace(/^claude-/, "Claude ")
    .replace(/^gpt-/, "GPT-")
    .replace(/-/g, " ")
    .replace(/\b\w/g, (c) => c.toUpperCase());
  return m;
}
