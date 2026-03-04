import { User } from 'lucide-react';
import { Card, CardContent } from '@/components/ui/card';

type AccountInfoProps = {
  userName: string;
  characterId: string;
};

export default function AccountInfo({ userName, characterId }: AccountInfoProps) {
  const portraitUrl = `https://image.eveonline.com/Character/${characterId}_128.jpg`;

  return (
    <Card>
      <CardContent className="p-6">
        <div className="flex items-center gap-2 mb-4">
          <User className="h-5 w-5 text-[var(--color-text-secondary)]" />
          <h3 className="text-lg font-semibold text-[var(--color-text-emphasis)]">Account</h3>
        </div>
        <div className="flex items-center gap-3">
          {/* eslint-disable-next-line @next/next/no-img-element */}
          <img
            src={portraitUrl}
            alt={userName}
            width={32}
            height={32}
            className="rounded-full"
          />
          <span className="text-sm font-medium text-[var(--color-text-emphasis)]">{userName}</span>
        </div>
      </CardContent>
    </Card>
  );
}
