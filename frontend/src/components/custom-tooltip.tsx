"use client";

interface PayloadEntry {
  name: string;
  value: number;
  color: string;
}

interface CustomTooltipProps {
  active?: boolean;
  payload?: PayloadEntry[];
  label?: string;
  formatter?: (value: number, name: string) => string;
}

export function CustomTooltip({
  active,
  payload,
  label,
  formatter,
}: CustomTooltipProps) {
  if (!active || !payload?.length) return null;

  return (
    <div className="bg-white border border-[var(--color-border-light)] rounded-lg px-4 py-3 text-[13px] shadow-[var(--shadow-tooltip)]">
      <div className="text-[var(--color-text-primary)] font-semibold mb-1.5">
        {label}
      </div>
      {payload.filter((p) => p.name !== "Trend").map((p, i) => (
        <div
          key={i}
          className="flex items-center gap-2 mt-0.5"
        >
          <span
            className="w-2.5 h-2.5 rounded-sm shrink-0"
            style={{ background: p.color }}
          />
          <span className="text-[var(--color-text-secondary)]">{p.name}:</span>
          <span className="text-[var(--color-text-primary)] font-medium">
            {formatter
              ? formatter(p.value, p.name)
              : p.value.toLocaleString()}
          </span>
        </div>
      ))}
    </div>
  );
}
