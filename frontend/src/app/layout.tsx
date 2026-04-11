import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: "Tirith Dashboard",
  description: "AI API Cost Observability",
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en">
      <head>
        <link
          href="https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700&display=swap"
          rel="stylesheet"
        />
        <link
          rel="stylesheet"
          href="https://use.typekit.net/qli0vdb.css"
        />
      </head>
      <body>{children}</body>
    </html>
  );
}
