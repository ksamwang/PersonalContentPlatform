const separators = /[,，、;；\n]+/;

export function normalizeTags(values: string[] | undefined) {
  if (!values) return [];
  const result: string[] = [];
  const seen = new Set<string>();
  for (const value of values.flatMap((item) => item.split(separators))) {
    const tag = value.trim();
    const key = tag.toLocaleLowerCase();
    if (!tag || seen.has(key)) continue;
    seen.add(key);
    result.push(tag);
  }
  return result;
}
