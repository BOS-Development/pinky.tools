import { Character } from "@industry-tool/client/data/models";
import Item from "./item";
import Navbar from "@industry-tool/components/Navbar";
import { Plus } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { AlertTriangle } from 'lucide-react';

export type CharacterListProps = {
  characters: Character[];
};

export default function List(props: CharacterListProps) {
  const reauthChars = props.characters.filter(c => c.needsReauth === true);

  if (props.characters.length == 0) {
    return (
      <>
        <Navbar />
        <div className="max-w-5xl mx-auto px-4 py-8">
          <div className="flex flex-col items-center justify-center min-h-[60vh] text-center">
            <h1 className="text-2xl font-display font-semibold mb-2">No Characters</h1>
            <p className="text-[var(--color-text-secondary)] mb-6">
              Get started by adding your first character
            </p>
            <Button size="lg" asChild>
              <a href="api/characters/add">
                <Plus className="h-4 w-4 mr-2" />
                Add Character
              </a>
            </Button>
          </div>
        </div>
      </>
    );
  }

  return (
    <>
      <Navbar />
      <div className="max-w-5xl mx-auto px-4 py-8">
        <div className="mb-6">
          <h1 className="text-2xl font-display font-semibold mb-3">Characters</h1>
          <Button asChild>
            <a href="api/characters/add">
              <Plus className="h-4 w-4 mr-2" />
              Add Character
            </a>
          </Button>
        </div>
        {reauthChars.map(char => (
          <Alert key={char.id} variant="destructive" className="mb-3 border-[var(--color-danger-rose)]/30 bg-[var(--color-danger-rose)]/10">
            <AlertTriangle className="h-4 w-4" />
            <AlertDescription className="flex items-center justify-between w-full">
              <span>
                ESI authorization for <strong>{char.name}</strong> has been revoked. Re-authorize to resume syncing.
              </span>
              <Button variant="outline" size="sm" asChild className="text-[var(--color-danger-rose)] border-[var(--color-danger-rose)]/30 ml-4">
                <a href="/api/characters/add">Re-authorize</a>
              </Button>
            </AlertDescription>
          </Alert>
        ))}
        <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 gap-4">
          {props.characters.map((char) => {
            return <Item character={char} key={char.id} />;
          })}
        </div>
      </div>
    </>
  );
}
