import type { Metadata } from "next";
import { AuthProvider } from "@/lib/auth-context";
import "./globals.css";

export const metadata: Metadata = {
  title: "satellite-tracker",
  description: "Dashboard for the satellite-tracker component & telemetry system",
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en">
      <body>
        <AuthProvider>
          <div className="shell">{children}</div>
        </AuthProvider>
      </body>
    </html>
  );
}
