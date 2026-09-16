type ArticleImage = {
  locale: string;
  publishedAt: string;
  slug: string;
  type: string;
};

export function socialImageURL(base: string, article?: ArticleImage) {
  const url = new URL("/api/share-image", base);
  if (article) {
    url.searchParams.set("locale", article.locale);
    url.searchParams.set("type", article.type);
    url.searchParams.set("slug", article.slug);
    url.searchParams.set("v", article.publishedAt);
  }
  return url.toString();
}
