import { Character, Corporation, AssetsResponse, StockpileMarker } from "./data/models";

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
  };
};

export default Client;
