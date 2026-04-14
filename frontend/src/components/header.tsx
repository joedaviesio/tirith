"use client";

interface Props {
  proxyActive: boolean;
}

export function Header({ proxyActive }: Props) {
  return (
    <header className="border-b border-[var(--color-border)] px-8 py-4 flex items-center justify-between">
      <div className="flex items-center gap-3">
        <span
          className="text-[36px] tracking-tight"
          style={{
            fontFamily: '"droog", sans-serif',
            fontWeight: 900,
            fontStyle: "normal",
            color: "var(--chart-line)",
          }}
        >
          Tirith
        </span>
        <span className="text-xs text-[var(--color-text-tertiary)] border border-[var(--color-border-light)] px-2 py-0.5 rounded-md">
          v2.0
        </span>
      </div>
      <div className="flex items-center gap-3">
        <div className="flex items-center gap-1.5">
          <span
            className="w-1.5 h-1.5 rounded-full"
            style={{
              background: proxyActive ? "#1d6bff" : "#cc0000",
            }}
          />
          <span
            className="snaga text-xs"
            style={{ color: proxyActive ? "#1d6bff" : "var(--color-text-secondary)" }}
          >
            {proxyActive ? "Proxy active on :5555" : "Proxy offline"}
          </span>
        </div>
      </div>
    </header>
  );
}
