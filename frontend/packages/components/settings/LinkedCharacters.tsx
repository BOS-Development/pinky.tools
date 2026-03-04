import { Users, Plus } from 'lucide-react';
import { Card, CardContent } from '@/components/ui/card';
import { Character } from '@industry-tool/client/data/models';

type LinkedCharactersProps = {
  characters: Character[];
};

export default function LinkedCharacters({ characters }: LinkedCharactersProps) {
  return (
    <Card>
      <CardContent className="p-6">
        <div className="flex items-center gap-2 mb-4">
          <Users className="h-5 w-5 text-[var(--color-primary-cyan)]" />
          <h3 className="text-lg font-semibold text-[var(--color-text-emphasis)]">Linked Characters</h3>
        </div>

        {characters.length === 0 ? (
          <p className="text-sm text-[var(--color-text-secondary)]">No characters linked yet.</p>
        ) : (
          <div className="space-y-2 mb-4">
            {characters.map((char) => (
              <div key={char.id} className="flex items-center gap-3">
                {/* eslint-disable-next-line @next/next/no-img-element */}
                <img
                  src={`https://image.eveonline.com/Character/${char.id}_128.jpg`}
                  alt={char.name}
                  width={32}
                  height={32}
                  className="rounded-full"
                />
                <span className="text-sm text-[var(--color-text-primary)]">{char.name}</span>
              </div>
            ))}
          </div>
        )}

        <a
          href="/api/characters/add"
          className="inline-flex items-center gap-1 text-sm text-[var(--color-primary-cyan)] hover:text-[var(--color-cyan-muted)] transition-colors"
        >
          <Plus className="h-4 w-4" />
          Add Character
        </a>
      </CardContent>
    </Card>
  );
}
