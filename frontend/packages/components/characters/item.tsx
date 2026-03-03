import { Character } from "@industry-tool/client/data/models";
import { characterScopesUpToDate } from "@industry-tool/client/scopes";
import { AlertTriangle, CircleAlert } from 'lucide-react';
import { Card, CardContent } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { cn } from '@/lib/utils';

export type CharacterItemProps = {
  character: Character;
};

export default function Item(props: CharacterItemProps) {
  const needsReauth = props.character.needsReauth === true;
  const needsScopeUpdate = !needsReauth && !characterScopesUpToDate(props.character);

  return (
    <Card className={cn(
      "max-w-[345px] transition-transform hover:-translate-y-1",
      needsReauth && "border-2 border-[var(--color-danger-rose)]",
      !needsReauth && needsScopeUpdate && "border-2 border-[var(--color-manufacturing-amber)]"
    )}>
      <div className="relative">
        <img
          src={`https://image.eveonline.com/Character/${props.character.id}_128.jpg`}
          alt={props.character.name}
          className="w-full max-w-[192px] aspect-square object-contain bg-[var(--color-bg-void)] mx-auto"
        />
        {needsReauth && (
          <div title="Authorization revoked — re-authorize required" className="absolute top-2 right-2">
            <CircleAlert className="h-7 w-7 text-[var(--color-danger-rose)] drop-shadow-lg" />
          </div>
        )}
        {needsScopeUpdate && (
          <div title="Scopes need updating" className="absolute top-2 right-2">
            <AlertTriangle className="h-7 w-7 text-[var(--color-manufacturing-amber)] drop-shadow-lg" />
          </div>
        )}
      </div>
      <CardContent className="p-4">
        <h3 className="text-lg font-semibold text-[var(--color-text-emphasis)]">
          {props.character.name}
        </h3>
        {needsReauth && (
          <div className="flex items-center gap-2 mt-2">
            <CircleAlert className="h-4 w-4 text-[var(--color-danger-rose)]" />
            <Button variant="outline" size="sm" asChild className="text-[var(--color-danger-rose)] border-[var(--color-danger-rose)]/30">
              <a href="/api/characters/add">Re-authorize</a>
            </Button>
          </div>
        )}
        {needsScopeUpdate && (
          <div className="flex items-center gap-2 mt-2">
            <AlertTriangle className="h-4 w-4 text-[var(--color-manufacturing-amber)]" />
            <Button variant="outline" size="sm" asChild className="text-[var(--color-manufacturing-amber)] border-[var(--color-manufacturing-amber)]/30">
              <a href="/api/characters/add">Re-authorize</a>
            </Button>
          </div>
        )}
      </CardContent>
    </Card>
  );
}
