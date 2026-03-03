import { Corporation } from "@industry-tool/client/data/models";
import Item from "./item";
import Navbar from "@industry-tool/components/Navbar";
import { Plus } from 'lucide-react';
import { Button } from '@/components/ui/button';

export type CorporationListProps = {
  corporations: Corporation[];
};

export default function List(props: CorporationListProps) {
  if (props.corporations.length == 0) {
    return (
      <>
        <Navbar />
        <div className="max-w-5xl mx-auto px-4 py-8">
          <div className="flex flex-col items-center justify-center min-h-[60vh] text-center">
            <h1 className="text-2xl font-display font-semibold mb-2">No Corporations</h1>
            <p className="text-[var(--color-text-secondary)] mb-6">
              Get started by adding your first corporation
            </p>
            <Button size="lg" asChild>
              <a href="api/corporations/add">
                <Plus className="h-4 w-4 mr-2" />
                Add Corporation
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
          <h1 className="text-2xl font-display font-semibold mb-3">Corporations</h1>
          <Button asChild>
            <a href="api/corporations/add">
              <Plus className="h-4 w-4 mr-2" />
              Add Corporation
            </a>
          </Button>
        </div>
        <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 gap-4">
          {props.corporations.map((corp) => {
            return <Item corporation={corp} key={corp.id} />;
          })}
        </div>
      </div>
    </>
  );
}
