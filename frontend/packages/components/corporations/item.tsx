import { Corporation } from "@industry-tool/client/data/models";
import { corporationScopesUpToDate } from "@industry-tool/client/scopes";
import { AlertTriangle, Building2 } from 'lucide-react';
import { Card, CardContent } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { cn } from '@/lib/utils';

export type CorporationItemProps = {
  corporation: Corporation;
};

export default function Item(props: CorporationItemProps) {
  const needsUpdate = !corporationScopesUpToDate(props.corporation);

  return (
    <Card className={cn(
      "max-w-[345px] transition-transform hover:-translate-y-1",
      needsUpdate && "border-2 border-[var(--color-manufacturing-amber)]"
    )}>
      <div className="h-[200px] flex items-center justify-center bg-[var(--color-bg-void)] relative overflow-hidden">
        <img
          src={`https://images.evetech.net/corporations/${props.corporation.id}/logo?size=256&tenant=tranquility`}
          alt={props.corporation.name}
          className="h-[200px] object-contain p-4 drop-shadow-lg"
          onError={(e) => {
            (e.target as HTMLImageElement).style.display = 'none';
          }}
        />
        {needsUpdate && (
          <div title="Scopes need updating" className="absolute top-2 right-2">
            <AlertTriangle className="h-7 w-7 text-[var(--color-manufacturing-amber)] drop-shadow-lg" />
          </div>
        )}
      </div>
      <CardContent className="p-4">
        <Badge variant="outline" className="mb-2">
          <Building2 className="h-3 w-3 mr-1" />
          Corporation
        </Badge>
        <h3 className="text-lg font-semibold text-[var(--color-primary-cyan)]">
          {props.corporation.name}
        </h3>
        {needsUpdate && (
          <div className="flex items-center gap-2 mt-2">
            <AlertTriangle className="h-4 w-4 text-[var(--color-manufacturing-amber)]" />
            <Button variant="outline" size="sm" asChild className="text-[var(--color-manufacturing-amber)] border-[var(--color-manufacturing-amber)]/30">
              <a href="/api/corporations/add">Re-authorize</a>
            </Button>
          </div>
        )}
      </CardContent>
    </Card>
  );
}
