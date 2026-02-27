import { characterScopes, corporationScopes } from "./scope-definitions";
import { Character, Corporation } from "./data/models";

export function characterScopesUpToDate(character: Character): boolean {
  if (!character.esiScopes) return false;
  const stored = new Set(character.esiScopes.split(" "));
  return characterScopes.every((scope) => stored.has(scope));
}

export function corporationScopesUpToDate(corporation: Corporation): boolean {
  if (!corporation.esiScopes) return false;
  const stored = new Set(corporation.esiScopes.split(" "));
  return corporationScopes.every((scope) => stored.has(scope));
}
