import { Asset, AssetStructure } from "@industry-tool/client/data/models";

export type AssetOwner = {
    ownerType: string;
    ownerId: number;
    ownerName: string;
};

export function getUniqueOwners(structure: AssetStructure): AssetOwner[] {
    const seen = new Map<string, AssetOwner>();

    const addOwner = (asset: Asset) => {
        const key = `${asset.ownerType}:${asset.ownerId}`;
        if (!seen.has(key)) {
            seen.set(key, { ownerType: asset.ownerType, ownerId: asset.ownerId, ownerName: asset.ownerName });
        }
    };

    for (const asset of structure.hangarAssets) addOwner(asset);
    for (const container of structure.hangarContainers) {
        for (const asset of container.assets) addOwner(asset);
    }
    for (const hanger of structure.corporationHangers) {
        for (const asset of hanger.assets) addOwner(asset);
        for (const container of hanger.hangarContainers) {
            for (const asset of container.assets) addOwner(asset);
        }
    }

    return Array.from(seen.values());
}

export function aggregateAssetsByTypeId(structure: AssetStructure): Map<number, number> {
    const totals = new Map<number, number>();

    const add = (typeId: number, qty: number) => {
        totals.set(typeId, (totals.get(typeId) || 0) + qty);
    };

    for (const asset of structure.hangarAssets) {
        add(asset.typeId, asset.quantity);
    }

    for (const container of structure.hangarContainers) {
        for (const asset of container.assets) {
            add(asset.typeId, asset.quantity);
        }
    }

    for (const asset of structure.deliveries) {
        add(asset.typeId, asset.quantity);
    }

    for (const asset of structure.assetSafety) {
        add(asset.typeId, asset.quantity);
    }

    for (const hanger of structure.corporationHangers) {
        for (const asset of hanger.assets) {
            add(asset.typeId, asset.quantity);
        }
        for (const container of hanger.hangarContainers) {
            for (const asset of container.assets) {
                add(asset.typeId, asset.quantity);
            }
        }
    }

    return totals;
}
