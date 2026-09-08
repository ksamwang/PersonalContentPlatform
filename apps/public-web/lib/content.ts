export type PublishedPage = {
  locale: string;
  type: string;
  slug: string;
  title: string;
  summary: string;
  html: string;
  published_at: string;
};
const origin = process.env.API_ORIGIN ?? "http://localhost:8080";
const workspace = process.env.WORKSPACE_SLUG ?? "personal";
export async function getPage(
  locale: string,
  type: string,
  slug: string,
): Promise<PublishedPage | null> {
  const response = await fetch(
    `${origin}/v1/public/${workspace}/${locale}/${type}/${slug}`,
    { next: { revalidate: 60, tags: [`${locale}:${type}:${slug}`] } },
  );
  if (response.status === 404) return null;
  if (!response.ok) throw new Error("Published content is unavailable");
  return response.json();
}
export async function listPages(locale: string): Promise<PublishedPage[]> {
  try {
    const response = await fetch(
      `${origin}/v1/public/${workspace}/${locale}/contents`,
      { next: { revalidate: 60, tags: [`${locale}:index`] } },
    );
    if (!response.ok) return [];
    return (await response.json()).items;
  } catch {
    return [];
  }
}
