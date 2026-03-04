'use client';
import { Toaster } from '@/components/ui/sonner';

// ThemeRegistry — simplified after MUI/Emotion removal.
// CSS variables in globals.css serve as the primary theme system for Tailwind/shadcn.
export default function ThemeRegistry({ children }: { children: React.ReactNode }) {
  return (
    <>
      {children}
      <Toaster position="bottom-center" />
    </>
  );
}
