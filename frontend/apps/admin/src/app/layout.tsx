import './globals.css';
import type { ReactNode } from 'react';

export const metadata = {
  title: 'Admin Dashboard',
  description: 'Provider analytics and administration'
};

export default function RootLayout({ children }: { children: ReactNode }) {
  return (
    <html lang="en">
      <body>{children}</body>
    </html>
  );
}
