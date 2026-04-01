"use client";

import { useState } from "react";
import { CallLog } from "@/lib/types";
import { getModelShortName } from "@/lib/constants";

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
          Recent calls
        </div>
        <div className="flex items-center justify-center py-12 text-[var(--color-text-tertiary)] text-sm">
          No API calls yet.
        </div>
      </div>
    );
  }

  function formatTime(ts: string): string {
    const d = new Date(ts);
    const now = new Date();
    const diffMs = now.getTime() - d.getTime();
    const diffMins = Math.floor(diffMs / 60000);

    if (diffMins < 1) return "Just now";
    if (diffMins < 60) return `${diffMins}m ago`;

    const diffHours = Math.floor(diffMins / 60);
    if (diffHours < 24) return `${diffHours}h ago`;

    return d.toLocaleDateString("en-US", {
      month: "short",
      day: "numeric",
      hour: "numeric",
      minute: "2-digit",
    });
  }

  function formatCost(cents: number): string {
    const dollars = cents / 100;
    if (dollars >= 1) return `$${dollars.toFixed(2)}`;
    if (dollars >= 0.01) return `$${dollars.toFixed(3)}`;
    if (dollars > 0) return `$${dollars.toFixed(4)}`;
    return "$0.00";
  }

  function formatLatency(ms: number): string {
    if (ms < 1000) return `${ms}ms`;
    return `${(ms / 1000).toFixed(1)}s`;
  }

  return (
    <div className="rounded-xl border border-[var(--color-border-light)] bg-[var(--color-surface)] shadow-[var(--shadow-card)]">
      {/* Header */}
      <div className="px-6 py-4 border-b border-[var(--color-border)]">
        <div className="text-[15px] font-semibold text-[var(--color-text-primary)]">
          Recent calls
        </div>
      </div>

      {/* Table */}
      <div className="overflow-x-auto">
        <table className="w-full border-collapse text-[13px]">
          <thead>
            <tr className="border-b border-[var(--color-border)]">
              {["WHEN", "MODEL", "COST", "SPEED"].map((h) => (
                <th
                  key={h}
                  className="px-6 py-3 text-left text-[11px] font-medium text-[var(--color-text-tertiary)] uppercase tracking-wider"
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
                <td className="px-6 py-3.5 text-[var(--color-text-secondary)] whitespace-nowrap">
                  {formatTime(c.timestamp)}
                </td>
                <td className="px-6 py-3.5 text-[var(--color-text-primary)] whitespace-nowrap">
                  {getModelShortName(c.model)}
                </td>
                <td className="px-6 py-3.5 text-[var(--color-text-primary)] font-tabular font-medium">
                  {formatCost(c.cost_cents)}
                </td>
                <td className="px-6 py-3.5 text-[var(--color-text-secondary)] font-tabular">
                  {formatLatency(c.latency_ms)}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {/* Pagination */}
      {totalPages > 1 && (
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
          <div className="text-xs text-[var(--color-text-tertiary)]">
            Page {page + 1} of {totalPages}
          </div>
        </div>
      )}
    </div>
  );
}
