import { Character, Corporation, AssetsResponse, StockpileMarker, HaulingRun, HaulingRunItem, HaulingArbitrageRow } from "./data/models";

export type ApiError = {
  kind: "error";
  statusCode: number;
  error: string;
};

export type ApiResult<T> = {
  kind: "success";
  data?: T;
};

export type FullCharacterData = {
  id: number;
  name: string;
  esiToken: string;
  esiRefreshToken: string;
  esiTokenExpiresOn: Date;
  esiScopes: string;
};

const getHeaders = (id: string) => {
  return {
    "Content-Type": "application/json",
    "USER-ID": id,
    "BACKEND-KEY": process.env.BACKEND_KEY as string,
  };
};

const Client = (baseUrl: string, id: string) => {
  return {
    async getCharacters(): Promise<ApiResult<Character[]> | ApiError> {
      let path = baseUrl + "v1/characters/";
      const response = await fetch(path, {
        method: "GET",
        headers: getHeaders(id),
      });

      if (response.status == 404) {
        return {
          kind: "error",
          statusCode: 404,
          error: "",
        };
      }

      if (response.status != 200) {
        throw `call to ${path} reponse code ${response.status}`;
      }

      const resp = await response.json();

      return {
        kind: "success",
        data: resp,
      };
    },
    async getCorporations(): Promise<ApiResult<Corporation[]> | ApiError> {
      let path = baseUrl + "v1/corporations";
      const response = await fetch(path, {
        method: "GET",
        headers: getHeaders(id),
      });

      if (response.status == 404) {
        return {
          kind: "error",
          statusCode: 404,
          error: "",
        };
      }

      if (response.status != 200) {
        throw `call to ${path} reponse code ${response.status}`;
      }

      const resp = await response.json();

      return {
        kind: "success",
        data: resp,
      };
    },
    async addCharacter(
      character: FullCharacterData,
    ): Promise<ApiResult<boolean> | ApiError> {
      const response = await fetch(baseUrl + "v1/characters/", {
        method: "POST",
        headers: getHeaders(id),
        body: JSON.stringify(character),
      });

      if (response.status != 200) {
        return {
          kind: "error",
          statusCode: response.status,
          error: "",
        };
      }

      return {
        kind: "success",
        data: true,
      };
    },
    async getAssetStatus(): Promise<ApiResult<{ lastUpdatedAt: string | null; nextUpdateAt: string | null }> | ApiError> {
      let path = baseUrl + "v1/users/asset-status";
      const response = await fetch(path, {
        method: "GET",
        headers: getHeaders(id),
      });

      if (response.status != 200) {
        throw `call to ${path} reponse code ${response.status}`;
      }

      const resp = await response.json();

      return {
        kind: "success",
        data: resp,
      };
    },
    async addCharacterCorporation(
      character: FullCharacterData,
    ): Promise<ApiResult<boolean> | ApiError> {
      const response = await fetch(baseUrl + "v1/corporations", {
        method: "POST",
        headers: getHeaders(id),
        body: JSON.stringify(character),
      });

      if (response.status != 200) {
        return {
          kind: "error",
          statusCode: response.status,
          error: "",
        };
      }

      return {
        kind: "success",
        data: true,
      };
    },
    async refreshStatic(): Promise<ApiResult<boolean> | ApiError> {
      let path = baseUrl + "v1/static/update";
      const response = await fetch(path, {
        method: "GET",
        headers: getHeaders(id),
      });

      if (response.status != 200) {
        throw `call to ${path} reponse code ${response.status}`;
      }

      return {
        kind: "success",
        data: true,
      };
    },
    async getAssets(): Promise<ApiResult<AssetsResponse> | ApiError> {
      let path = baseUrl + "v1/assets/";
      const response = await fetch(path, {
        method: "GET",
        headers: getHeaders(id),
      });

      if (response.status == 404) {
        return {
          kind: "error",
          statusCode: 404,
          error: "",
        };
      }

      if (response.status != 200) {
        throw `call to ${path} reponse code ${response.status}`;
      }

      const resp = await response.json();

      return {
        kind: "success",
        data: resp,
      };
    },
    async getStockpiles(): Promise<ApiResult<StockpileMarker[]> | ApiError> {
      let path = baseUrl + "v1/stockpiles";
      const response = await fetch(path, {
        method: "GET",
        headers: getHeaders(id),
      });

      if (response.status != 200) {
        throw `call to ${path} response code ${response.status}`;
      }

      const resp = await response.json();

      return {
        kind: "success",
        data: resp,
      };
    },
    async upsertStockpile(
      marker: StockpileMarker
    ): Promise<ApiResult<boolean> | ApiError> {
      const response = await fetch(baseUrl + "v1/stockpiles", {
        method: "POST",
        headers: getHeaders(id),
        body: JSON.stringify(marker),
      });

      if (response.status != 200) {
        return {
          kind: "error",
          statusCode: response.status,
          error: "",
        };
      }

      return {
        kind: "success",
        data: true,
      };
    },
    async deleteStockpile(
      marker: StockpileMarker
    ): Promise<ApiResult<boolean> | ApiError> {
      const response = await fetch(baseUrl + "v1/stockpiles", {
        method: "DELETE",
        headers: getHeaders(id),
        body: JSON.stringify(marker),
      });

      if (response.status != 200) {
        return {
          kind: "error",
          statusCode: response.status,
          error: "",
        };
      }

      return {
        kind: "success",
        data: true,
      };
    },

    async getHaulingRuns(): Promise<ApiResult<HaulingRun[]> | ApiError> {
      const path = baseUrl + "v1/hauling/runs";
      const response = await fetch(path, {
        method: "GET",
        headers: getHeaders(id),
      });

      if (response.status !== 200) {
        throw `call to ${path} response code ${response.status}`;
      }

      const resp = await response.json();
      return { kind: "success", data: resp };
    },

    async createHaulingRun(run: Partial<HaulingRun>): Promise<ApiResult<HaulingRun> | ApiError> {
      const path = baseUrl + "v1/hauling/runs";
      const response = await fetch(path, {
        method: "POST",
        headers: getHeaders(id),
        body: JSON.stringify(run),
      });

      if (response.status !== 200 && response.status !== 201) {
        return { kind: "error", statusCode: response.status, error: "" };
      }

      const resp = await response.json();
      return { kind: "success", data: resp };
    },

    async getHaulingRun(runId: number): Promise<ApiResult<HaulingRun> | ApiError> {
      const path = baseUrl + `v1/hauling/runs/${runId}`;
      const response = await fetch(path, {
        method: "GET",
        headers: getHeaders(id),
      });

      if (response.status === 404) {
        return { kind: "error", statusCode: 404, error: "" };
      }

      if (response.status !== 200) {
        throw `call to ${path} response code ${response.status}`;
      }

      const resp = await response.json();
      return { kind: "success", data: resp };
    },

    async updateHaulingRun(runId: number, run: Partial<HaulingRun>): Promise<ApiResult<HaulingRun> | ApiError> {
      const path = baseUrl + `v1/hauling/runs/${runId}`;
      const response = await fetch(path, {
        method: "PUT",
        headers: getHeaders(id),
        body: JSON.stringify(run),
      });

      if (response.status !== 200) {
        return { kind: "error", statusCode: response.status, error: "" };
      }

      const resp = await response.json();
      return { kind: "success", data: resp };
    },

    async deleteHaulingRun(runId: number): Promise<ApiResult<boolean> | ApiError> {
      const path = baseUrl + `v1/hauling/runs/${runId}`;
      const response = await fetch(path, {
        method: "DELETE",
        headers: getHeaders(id),
      });

      if (response.status !== 200 && response.status !== 204) {
        return { kind: "error", statusCode: response.status, error: "" };
      }

      return { kind: "success", data: true };
    },

    async updateHaulingRunStatus(runId: number, status: string): Promise<ApiResult<HaulingRun> | ApiError> {
      const path = baseUrl + `v1/hauling/runs/${runId}/status`;
      const response = await fetch(path, {
        method: "PUT",
        headers: getHeaders(id),
        body: JSON.stringify({ status }),
      });

      if (response.status !== 200) {
        return { kind: "error", statusCode: response.status, error: "" };
      }

      const resp = await response.json();
      return { kind: "success", data: resp };
    },

    async addHaulingRunItem(runId: number, item: Partial<HaulingRunItem>): Promise<ApiResult<HaulingRunItem> | ApiError> {
      const path = baseUrl + `v1/hauling/runs/${runId}/items`;
      const response = await fetch(path, {
        method: "POST",
        headers: getHeaders(id),
        body: JSON.stringify(item),
      });

      if (response.status !== 200 && response.status !== 201) {
        return { kind: "error", statusCode: response.status, error: "" };
      }

      const resp = await response.json();
      return { kind: "success", data: resp };
    },

    async updateHaulingRunItemAcquired(runId: number, itemId: number, quantityAcquired: number): Promise<ApiResult<HaulingRunItem> | ApiError> {
      const path = baseUrl + `v1/hauling/runs/${runId}/items/${itemId}`;
      const response = await fetch(path, {
        method: "PUT",
        headers: getHeaders(id),
        body: JSON.stringify({ quantityAcquired }),
      });

      if (response.status !== 200) {
        return { kind: "error", statusCode: response.status, error: "" };
      }

      const resp = await response.json();
      return { kind: "success", data: resp };
    },

    async removeHaulingRunItem(runId: number, itemId: number): Promise<ApiResult<boolean> | ApiError> {
      const path = baseUrl + `v1/hauling/runs/${runId}/items/${itemId}`;
      const response = await fetch(path, {
        method: "DELETE",
        headers: getHeaders(id),
      });

      if (response.status !== 200 && response.status !== 204) {
        return { kind: "error", statusCode: response.status, error: "" };
      }

      return { kind: "success", data: true };
    },

    async getScannerResults(sourceRegionId: number, destRegionId: number, sourceSystemId?: number): Promise<ApiResult<HaulingArbitrageRow[]> | ApiError> {
      const params = new URLSearchParams({
        source_region_id: String(sourceRegionId),
        dest_region_id: String(destRegionId),
      });
      if (sourceSystemId) params.set("source_system_id", String(sourceSystemId));

      const path = baseUrl + `v1/hauling/scanner?${params.toString()}`;
      const response = await fetch(path, {
        method: "GET",
        headers: getHeaders(id),
      });

      if (response.status !== 200) {
        throw `call to ${path} response code ${response.status}`;
      }

      const resp = await response.json();
      return { kind: "success", data: resp };
    },

    async triggerHaulingScan(regionId: number, systemId?: number): Promise<ApiResult<boolean> | ApiError> {
      const path = baseUrl + "v1/hauling/scanner/scan";
      const body: Record<string, number> = { regionId };
      if (systemId) body.systemId = systemId;

      const response = await fetch(path, {
        method: "POST",
        headers: getHeaders(id),
        body: JSON.stringify(body),
      });

      if (response.status !== 200 && response.status !== 202) {
        return { kind: "error", statusCode: response.status, error: "" };
      }

      return { kind: "success", data: true };
    },
  };
};

export default Client;
