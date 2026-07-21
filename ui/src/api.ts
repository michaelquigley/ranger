import createClient from "openapi-fetch";
import type { components, paths } from "./api/schema";

export type Board = components["schemas"]["board"];
export type Lane = components["schemas"]["lane"];
export type Card = components["schemas"]["card"];
export type State = components["schemas"]["state"];
export type Conflict = components["schemas"]["conflict"];
export type ProjectIndex = components["schemas"]["projectIndex"];
export type ProjectStatus = components["schemas"]["projectStatus"];

export const client = createClient<paths>({ baseUrl: "/api/v1" });

// every mutation resolves to one of four outcomes the board knows how to
// handle: a fresh board (with the landing filename for the rename family),
// a typed conflict, a validation refusal, or a server fault whose message
// is surfaced verbatim.
export type Outcome =
  | { kind: "ok"; board: Board; filename?: string }
  | { kind: "conflict"; conflict: Conflict }
  | { kind: "invalid"; message: string }
  | { kind: "fault"; message: string };

function failure(status: number, error: unknown): Outcome {
  const body = error as { reason?: string; message?: string } | undefined;
  if (status === 409 && body?.reason) {
    return { kind: "conflict", conflict: body as Conflict };
  }
  if (status === 400) {
    return { kind: "invalid", message: body?.message ?? "invalid request" };
  }
  return { kind: "fault", message: body?.message ?? `request failed (${status})` };
}

export type ItemDetail = {
  content: string;
  card: Card;
  hash: string;
};

// fetchProjects is the one unscoped call: the index the selector renders
// and the bare / consults for its redirect.
export async function fetchProjects(): Promise<ProjectIndex> {
  const { data, error, response } = await client.GET("/projects");
  if (data) return data;
  const body = error as { message?: string } | undefined;
  throw new Error(body?.message ?? `project index failed (${response.status})`);
}

// makeApi binds every operation to one project — the one UI seam. call
// sites never thread the name; everything downstream keys off the URL's
// project through this binding.
export function makeApi(project: string) {
  return {
    project,

    async fetchBoard(): Promise<Board> {
      const { data, error, response } = await client.GET("/projects/{project}/board", {
        params: { path: { project } },
      });
      if (data) return data;
      const body = error as { message?: string } | undefined;
      throw new Error(body?.message ?? `board load failed (${response.status})`);
    },

    async fetchItem(filename: string): Promise<ItemDetail> {
      const { data, error, response } = await client.GET("/projects/{project}/items/{filename}", {
        params: { path: { project, filename } },
      });
      if (data) return data;
      const body = error as { message?: string } | undefined;
      throw new Error(body?.message ?? `item load failed (${response.status})`);
    },

    async search(q: string): Promise<string[]> {
      const { data, error, response } = await client.GET("/projects/{project}/search", {
        params: { path: { project }, query: { q } },
      });
      if (data) return data.filenames;
      const body = error as { message?: string } | undefined;
      throw new Error(body?.message ?? `search failed (${response.status})`);
    },

    async capture(title: string, body: string): Promise<Outcome> {
      const { data, error, response } = await client.POST("/projects/{project}/items", {
        params: { path: { project } },
        body: body ? { title, body } : { title },
      });
      if (data) return { kind: "ok", board: data.board, filename: data.filename };
      return failure(response.status, error);
    },

    async transition(
      filename: string,
      state: State,
      expectedHash: string,
      expectedOrderVersion: string,
      position?: number,
    ): Promise<Outcome> {
      const { data, error, response } = await client.POST("/projects/{project}/items/{filename}/state", {
        params: { path: { project, filename } },
        body: { state, expectedHash, expectedOrderVersion, ...(position !== undefined ? { position } : {}) },
      });
      if (data) return { kind: "ok", board: data };
      return failure(response.status, error);
    },

    async reorder(lane: State, filenames: string[], expectedVersion: string): Promise<Outcome> {
      const { data, error, response } = await client.PUT("/projects/{project}/order/{lane}", {
        params: { path: { project, lane } },
        body: { filenames, expectedVersion },
      });
      if (data) return { kind: "ok", board: data };
      return failure(response.status, error);
    },

    async saveContent(
      filename: string,
      content: string,
      expectedHash: string,
      expectedOrderVersion: string,
    ): Promise<Outcome> {
      const { data, error, response } = await client.PUT("/projects/{project}/items/{filename}/content", {
        params: { path: { project, filename } },
        body: { content, expectedHash, expectedOrderVersion },
      });
      if (data) return { kind: "ok", board: data };
      return failure(response.status, error);
    },

    async retitle(
      filename: string,
      title: string,
      expectedHash: string,
      expectedOrderVersion: string,
    ): Promise<Outcome> {
      const { data, error, response } = await client.POST("/projects/{project}/items/{filename}/retitle", {
        params: { path: { project, filename } },
        body: { title, expectedHash, expectedOrderVersion },
      });
      if (data) return { kind: "ok", board: data.board, filename: data.filename };
      return failure(response.status, error);
    },

    async deleteItem(filename: string, expectedHash: string, expectedOrderVersion: string): Promise<Outcome> {
      const { data, error, response } = await client.POST("/projects/{project}/items/{filename}/delete", {
        params: { path: { project, filename } },
        body: { expectedHash, expectedOrderVersion },
      });
      if (data) return { kind: "ok", board: data };
      return failure(response.status, error);
    },

    async renameToSlug(filename: string, expectedHash: string, expectedOrderVersion: string): Promise<Outcome> {
      const { data, error, response } = await client.POST("/projects/{project}/items/{filename}/rename-to-slug", {
        params: { path: { project, filename } },
        body: { expectedHash, expectedOrderVersion },
      });
      if (data) return { kind: "ok", board: data.board, filename: data.filename };
      return failure(response.status, error);
    },
  };
}

export type Api = ReturnType<typeof makeApi>;
