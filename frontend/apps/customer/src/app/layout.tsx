import './globals.css';
import type { ReactNode } from 'react';

export const metadata = {
  title: 'Customer Portal',
  description: 'Usage, quota, and billing overview'
};

export default function RootLayout({ children }: { children: ReactNode }) {
  return (
    <html lang="en">
      <body>{children}</body>
    </html>
  );
}
