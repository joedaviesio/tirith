"use client";

import { useState, useRef, useEffect } from "react";

interface Props {
  models: string[];
  selectedModel: string;
  onModelChange: (model: string) => void;
  timeRange: string;
  onTimeRangeChange: (range: string) => void;
}

const TIME_RANGES = [
  { value: "1h", label: "Last hour" },
  { value: "24h", label: "Today" },
  { value: "7d", label: "7 days" },
  { value: "30d", label: "30 days" },
];

export function FilterBar({
  models,
  selectedModel,
  onModelChange,
  timeRange,
  onTimeRangeChange,
}: Props) {
  const [modelOpen, setModelOpen] = useState(false);
  const [modelSearch, setModelSearch] = useState("");
  const modelRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    function handleClick(e: MouseEvent) {
      if (modelRef.current && !modelRef.current.contains(e.target as Node)) {
        setModelOpen(false);
      }
    }
    document.addEventListener("mousedown", handleClick);
    return () => document.removeEventListener("mousedown", handleClick);
  }, []);

  const filteredModels = models.filter((m) =>
    m.toLowerCase().includes(modelSearch.toLowerCase())
  );

  return (
    <div className="flex items-center justify-between flex-wrap gap-3">
      {/* Time Range */}
      <div className="flex items-center gap-1">
        {TIME_RANGES.map((tr) => (
          <button
            key={tr.value}
            onClick={() => onTimeRangeChange(tr.value)}
            className="snaga px-4 py-2 rounded-full text-[14px] font-medium cursor-pointer transition-all"
            style={{
              border:
                timeRange === tr.value
                  ? "1.5px solid var(--color-border-active)"
                  : "1.5px solid transparent",
              background:
                timeRange === tr.value
                  ? "var(--color-surface)"
                  : "transparent",
              color:
                timeRange === tr.value
                  ? "var(--color-text-primary)"
                  : "var(--color-text-secondary)",
            }}
          >
            {tr.label}
          </button>
        ))}
      </div>

      {/* Model Filter */}
      <div className="relative" ref={modelRef}>
        <button
          onClick={() => setModelOpen(!modelOpen)}
          className="flex items-center gap-1.5 px-3 py-1.5 rounded-full text-[13px] font-medium cursor-pointer border-[1.5px] border-[var(--color-border-light)] bg-[var(--color-surface)]"
        >
          <span className="snaga text-[var(--color-text-secondary)]">Model:</span>
          <span className="text-[var(--color-text-primary)]">
            {selectedModel || "All"}
          </span>
          <svg
            width="12"
            height="12"
            viewBox="0 0 12 12"
            fill="none"
            className="text-[var(--color-text-tertiary)]"
          >
            <path
              d="M3 4.5L6 7.5L9 4.5"
              stroke="currentColor"
              strokeWidth="1.5"
              strokeLinecap="round"
              strokeLinejoin="round"
            />
          </svg>
        </button>

        {modelOpen && (
          <div className="absolute top-full mt-1 right-0 w-64 bg-[var(--color-surface)] border border-[var(--color-border-light)] rounded-lg shadow-[var(--shadow-tooltip)] z-50">
            <div className="p-2">
              <input
                type="text"
                placeholder="Search models..."
                value={modelSearch}
                onChange={(e) => setModelSearch(e.target.value)}
                className="w-full px-3 py-1.5 text-[13px] border border-[var(--color-border)] rounded-md bg-[var(--color-surface)] text-[var(--color-text-primary)] outline-none focus:border-[var(--color-border-active)]"
                autoFocus
              />
            </div>
            <div className="max-h-48 overflow-y-auto">
              <button
                onClick={() => {
                  onModelChange("");
                  setModelOpen(false);
                  setModelSearch("");
                }}
                className="w-full text-left px-4 py-2 text-[13px] hover:bg-[var(--color-surface-alt)] cursor-pointer"
                style={{
                  fontWeight: selectedModel === "" ? 700 : 400,
                  color: "var(--color-text-primary)",
                }}
              >
                All models
              </button>
              {filteredModels.map((m) => (
                <button
                  key={m}
                  onClick={() => {
                    onModelChange(m);
                    setModelOpen(false);
                    setModelSearch("");
                  }}
                  className="w-full text-left px-4 py-2 text-[13px] hover:bg-[var(--color-surface-alt)] cursor-pointer"
                  style={{
                    fontWeight: selectedModel === m ? 700 : 400,
                    color: "var(--color-text-primary)",
                  }}
                >
                  {m}
                </button>
              ))}
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
