import { Lock } from 'lucide-react';
import { Card, CardContent } from '@/components/ui/card';
import { Button } from '@/components/ui/button';

export default function Unauthorized() {
  return (
    <div className="flex items-center justify-center min-h-screen bg-[var(--color-bg-void)]">
      <Card className="max-w-[400px] w-full">
        <CardContent className="flex flex-col items-center text-center p-8">
          <Lock className="h-14 w-14 mb-4 text-[var(--color-primary-cyan)]" />
          <h2 className="text-xl font-semibold text-[var(--color-text-emphasis)] mb-2">
            Authentication Required
          </h2>
          <p className="text-sm text-[var(--color-text-secondary)] mb-6">
            You must sign in to access this page
          </p>
          <Button asChild size="lg" className="w-full">
            <a href="api/auth/login">Sign In</a>
          </Button>
        </CardContent>
      </Card>
    </div>
  );
}
