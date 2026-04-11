interface StatCardProps {
  label: string;
  value: string | number;
  sub?: string;
}

export function StatCard({ label, value, sub }: StatCardProps) {
  return (
    <div className="rounded-xl border border-[var(--color-border-light)] bg-[var(--color-surface)] p-5 flex-1 min-w-[160px] shadow-[var(--shadow-card)]">
      <div className="text-[13px] text-[var(--color-text-secondary)] mb-1">
        {label}
      </div>
      <div className="text-[28px] font-bold text-[var(--color-text-primary)] font-tabular tracking-tight">
        {value}
      </div>
      {sub && (
        <div className="text-xs text-[var(--color-text-tertiary)] mt-1">
          {sub}
        </div>
      )}
    </div>
  );
}
