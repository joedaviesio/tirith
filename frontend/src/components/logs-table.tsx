"use client";

import { useState } from "react";
import { CallLog } from "@/lib/types";

interface Props {
  calls: CallLog[];
}

const PAGE_SIZE = 10;

export function LogsTable({ calls }: Props) {
  const [page, setPage] = useState(0);
  const totalPages = Math.ceil(calls.length / PAGE_SIZE);
  const paginated = calls.slice(page * PAGE_SIZE, (page + 1) * PAGE_SIZE);

  if (calls.length === 0) {
    return (
      <div className="rounded-xl border border-[var(--color-border-light)] bg-[var(--color-surface)] p-6 shadow-[var(--shadow-card)]">
        <div className="text-[15px] font-semibold text-[var(--color-text-primary)] mb-4">
          Logs
        </div>
        <div className="flex items-center justify-center py-12 text-[var(--color-text-tertiary)] text-sm">
          No API calls logged yet.
        </div>
      </div>
    );
  }

  function formatTimestamp(ts: string): string {
    const d = new Date(ts);
    return d.toLocaleString("en-US", {
      year: "numeric",
      month: "2-digit",
      day: "2-digit",
      hour: "2-digit",
      minute: "2-digit",
      second: "2-digit",
      hour12: false,
    });
  }

  function truncateId(id: string): string {
    if (id.length <= 20) return id;
    return id.slice(0, 20) + "...";
  }

  return (
    <div className="rounded-xl border border-[var(--color-border-light)] bg-[var(--color-surface)] shadow-[var(--shadow-card)]">
      {/* Header */}
      <div className="px-6 py-4 flex items-center justify-between border-b border-[var(--color-border)]">
        <div className="text-[15px] font-semibold text-[var(--color-text-primary)]">
          Logs
        </div>
        <div className="text-xs text-[var(--color-text-tertiary)]">
          Last refresh: {new Date().toLocaleString()}
        </div>
      </div>

      {/* Table */}
      <div className="overflow-x-auto">
        <table className="w-full border-collapse text-[13px]">
          <thead>
            <tr className="border-b border-[var(--color-border)]">
              {[
                "TIME",
                "ID",
                "MODEL",
                "INPUT TOKENS",
                "OUTPUT TOKENS",
                "TYPE",
                "SERVICE TIER",
                "REQUEST",
              ].map((h) => (
                <th
                  key={h}
                  className="px-4 py-3 text-left text-[11px] font-medium text-[var(--color-text-tertiary)] uppercase tracking-wider"
                >
                  {h}
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            {paginated.map((c) => (
              <tr
                key={c.id}
                className="border-b border-[var(--color-border-light)] hover:bg-[var(--color-surface-alt)] transition-colors"
              >
                <td className="px-4 py-3 text-[var(--color-text-secondary)] whitespace-nowrap font-tabular">
                  {formatTimestamp(c.timestamp)}
                </td>
                <td className="px-4 py-3 text-[var(--color-text-secondary)] font-mono text-xs">
                  {truncateId(c.id)}
                </td>
                <td className="px-4 py-3 text-[var(--color-text-primary)] whitespace-nowrap">
                  {c.model}
                </td>
                <td className="px-4 py-3 text-[var(--color-text-primary)] font-tabular">
                  {c.input_tokens.toLocaleString()}
                </td>
                <td className="px-4 py-3 text-[var(--color-text-primary)] font-tabular">
                  {c.output_tokens.toLocaleString()}
                </td>
                <td className="px-4 py-3 text-[var(--color-text-secondary)]">
                  {c.streaming ? "Streaming" : "HTTP"}
                </td>
                <td className="px-4 py-3 text-[var(--color-text-secondary)]">
                  Standard
                </td>
                <td className="px-4 py-3 text-[var(--color-text-secondary)]">
                  {c.streaming ? "Streaming" : "HTTP"}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {/* Pagination */}
      <div className="px-6 py-3 flex items-center justify-between border-t border-[var(--color-border)]">
        <div className="flex items-center gap-2">
          <button
            onClick={() => setPage(Math.max(0, page - 1))}
            disabled={page === 0}
            className="w-8 h-8 flex items-center justify-center rounded-md border border-[var(--color-border-light)] text-[var(--color-text-secondary)] disabled:opacity-30 cursor-pointer disabled:cursor-default hover:bg-[var(--color-surface-alt)]"
          >
            &larr;
          </button>
          <button
            onClick={() => setPage(Math.min(totalPages - 1, page + 1))}
            disabled={page >= totalPages - 1}
            className="w-8 h-8 flex items-center justify-center rounded-md border border-[var(--color-border-light)] text-[var(--color-text-secondary)] disabled:opacity-30 cursor-pointer disabled:cursor-default hover:bg-[var(--color-surface-alt)]"
          >
            &rarr;
          </button>
        </div>
        <div className="flex items-center gap-2 text-xs text-[var(--color-text-tertiary)]">
          <span>
            LINES PER PAGE
          </span>
          <span className="font-medium text-[var(--color-text-primary)]">
            {PAGE_SIZE}
          </span>
        </div>
      </div>
    </div>
  );
}
