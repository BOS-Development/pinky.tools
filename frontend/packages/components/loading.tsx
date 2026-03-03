import { Loader2 } from 'lucide-react';

export default function Loading() {
  return (
    <div className="flex flex-col items-center justify-center min-h-screen gap-3">
      <Loader2 className="h-12 w-12 animate-spin text-[var(--color-primary-cyan)]" />
      <p className="text-base text-[var(--color-text-secondary)]">
        Loading...
      </p>
    </div>
  );
}
