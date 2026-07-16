import createClient from "openapi-fetch";
import type { components, paths } from "./api/schema";

export type Board = components["schemas"]["board"];
export type Lane = components["schemas"]["lane"];
export type Card = components["schemas"]["card"];
export type State = components["schemas"]["state"];
export type Conflict = components["schemas"]["conflict"];

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

export async function fetchBoard(): Promise<Board> {
  const { data, error, response } = await client.GET("/board");
  if (data) return data;
  const body = error as { message?: string } | undefined;
  throw new Error(body?.message ?? `board load failed (${response.status})`);
}

export type ItemDetail = {
  content: string;
  card: Card;
  hash: string;
};

export async function fetchItem(filename: string): Promise<ItemDetail> {
  const { data, error, response } = await client.GET("/items/{filename}", {
    params: { path: { filename } },
  });
  if (data) return data;
  const body = error as { message?: string } | undefined;
  throw new Error(body?.message ?? `item load failed (${response.status})`);
}

export async function capture(title: string, body: string): Promise<Outcome> {
  const { data, error, response } = await client.POST("/items", {
    body: body ? { title, body } : { title },
  });
  if (data) return { kind: "ok", board: data.board, filename: data.filename };
  return failure(response.status, error);
}

export async function transition(
  filename: string,
  state: State,
  expectedHash: string,
  expectedOrderVersion: string,
  position?: number,
): Promise<Outcome> {
  const { data, error, response } = await client.POST("/items/{filename}/state", {
    params: { path: { filename } },
    body: { state, expectedHash, expectedOrderVersion, ...(position !== undefined ? { position } : {}) },
  });
  if (data) return { kind: "ok", board: data };
  return failure(response.status, error);
}

export async function reorder(lane: State, filenames: string[], expectedVersion: string): Promise<Outcome> {
  const { data, error, response } = await client.PUT("/order/{lane}", {
    params: { path: { lane } },
    body: { filenames, expectedVersion },
  });
  if (data) return { kind: "ok", board: data };
  return failure(response.status, error);
}

export async function saveContent(
  filename: string,
  content: string,
  expectedHash: string,
  expectedOrderVersion: string,
): Promise<Outcome> {
  const { data, error, response } = await client.PUT("/items/{filename}/content", {
    params: { path: { filename } },
    body: { content, expectedHash, expectedOrderVersion },
  });
  if (data) return { kind: "ok", board: data };
  return failure(response.status, error);
}

export async function retitle(
  filename: string,
  title: string,
  expectedHash: string,
  expectedOrderVersion: string,
): Promise<Outcome> {
  const { data, error, response } = await client.POST("/items/{filename}/retitle", {
    params: { path: { filename } },
    body: { title, expectedHash, expectedOrderVersion },
  });
  if (data) return { kind: "ok", board: data.board, filename: data.filename };
  return failure(response.status, error);
}

export async function deleteItem(
  filename: string,
  expectedHash: string,
  expectedOrderVersion: string,
): Promise<Outcome> {
  const { data, error, response } = await client.POST("/items/{filename}/delete", {
    params: { path: { filename } },
    body: { expectedHash, expectedOrderVersion },
  });
  if (data) return { kind: "ok", board: data };
  return failure(response.status, error);
}

export async function renameToSlug(
  filename: string,
  expectedHash: string,
  expectedOrderVersion: string,
): Promise<Outcome> {
  const { data, error, response } = await client.POST("/items/{filename}/rename-to-slug", {
    params: { path: { filename } },
    body: { expectedHash, expectedOrderVersion },
  });
  if (data) return { kind: "ok", board: data.board, filename: data.filename };
  return failure(response.status, error);
}
