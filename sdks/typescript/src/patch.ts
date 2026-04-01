/**
 * Monkey-patching logic for Anthropic and OpenAI SDKs.
 */

const PATCHED_MARKER = Symbol("tirith_patched");

export async function patchAnthropic(proxyUrl: string): Promise<void> {
  let anthropicMod: any;
  try {
    anthropicMod = await import("@anthropic-ai/sdk");
  } catch {
    // @anthropic-ai/sdk not installed — skip.
    return;
  }

  const AnthropicClass = anthropicMod.default || anthropicMod.Anthropic;
  if (!AnthropicClass || (AnthropicClass as any)[PATCHED_MARKER]) {
    return;
  }

  const anthropicBase = `${proxyUrl}/proxy/anthropic`;
  const OriginalConstructor = AnthropicClass;

  // Create a subclass that defaults baseURL.
  const PatchedAnthropic = class extends OriginalConstructor {
    constructor(opts: any = {}) {
      opts.baseURL ??= anthropicBase;
      super(opts);
    }
  };

  (PatchedAnthropic as any)[PATCHED_MARKER] = true;

  // Replace on the module.
  if (anthropicMod.default) {
    anthropicMod.default = PatchedAnthropic;
  }
  if (anthropicMod.Anthropic) {
    anthropicMod.Anthropic = PatchedAnthropic;
  }

  // Patch AsyncAnthropic if it exists.
  const AsyncAnthropicClass = anthropicMod.AsyncAnthropic;
  if (AsyncAnthropicClass && !(AsyncAnthropicClass as any)[PATCHED_MARKER]) {
    const PatchedAsync = class extends AsyncAnthropicClass {
      constructor(opts: any = {}) {
        opts.baseURL ??= anthropicBase;
        super(opts);
      }
    };
    (PatchedAsync as any)[PATCHED_MARKER] = true;
    anthropicMod.AsyncAnthropic = PatchedAsync;
  }
}

export async function patchOpenAI(proxyUrl: string): Promise<void> {
  let openaiMod: any;
  try {
    openaiMod = await import("openai");
  } catch {
    // openai not installed — skip.
    return;
  }

  const OpenAIClass = openaiMod.default || openaiMod.OpenAI;
  if (!OpenAIClass || (OpenAIClass as any)[PATCHED_MARKER]) {
    return;
  }

  const openaiBase = `${proxyUrl}/proxy/openai`;
  const OriginalConstructor = OpenAIClass;

  const PatchedOpenAI = class extends OriginalConstructor {
    constructor(opts: any = {}) {
      opts.baseURL ??= openaiBase;
      super(opts);
    }
  };

  (PatchedOpenAI as any)[PATCHED_MARKER] = true;

  if (openaiMod.default) {
    openaiMod.default = PatchedOpenAI;
  }
  if (openaiMod.OpenAI) {
    openaiMod.OpenAI = PatchedOpenAI;
  }
}
