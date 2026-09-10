import { revalidateTag } from "next/cache";
import { NextRequest, NextResponse } from "next/server";

export async function POST(request: NextRequest) {
  const expected = process.env.PUBLIC_REVALIDATE_TOKEN?.trim();
  if (expected && request.headers.get("authorization") !== `Bearer ${expected}`) {
    return NextResponse.json({ error: "unauthorized" }, { status: 401 });
  }
  const body = await request.json().catch(() => ({})) as { workspace_id?: string };
  if (!body.workspace_id || !/^[0-9a-f-]{36}$/i.test(body.workspace_id)) {
    return NextResponse.json({ error: "workspace_id is required" }, { status: 400 });
  }
  revalidateTag(`pcp:${body.workspace_id}`, "max");
  return NextResponse.json({ revalidated: true });
}
