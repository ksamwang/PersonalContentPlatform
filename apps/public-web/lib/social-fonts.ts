import { readFile } from "node:fs/promises";
import { join } from "node:path";

type SocialFont = {
  data: ArrayBuffer;
  name: string;
  style: "normal";
  weight: 400 | 700;
};

function arrayBuffer(data: Buffer) {
  return data.buffer.slice(
    data.byteOffset,
    data.byteOffset + data.byteLength,
  ) as ArrayBuffer;
}

let fontsPromise: Promise<SocialFont[]> | undefined;

export function socialFonts() {
  fontsPromise ??= Promise.all([
    readFile(join(process.cwd(), "assets", "fonts", "PublicSans-Regular.ttf")),
    readFile(join(process.cwd(), "assets", "fonts", "PublicSans-Bold.ttf")),
    readFile(join(process.cwd(), "assets", "fonts", "Newsreader-Regular.ttf")),
    readFile(join(process.cwd(), "assets", "fonts", "Newsreader-Bold.ttf")),
    readFile(join(process.cwd(), "assets", "fonts", "NotoSansSC-Regular.otf")),
    readFile(join(process.cwd(), "assets", "fonts", "NotoSansSC-Bold.otf")),
  ]).then(
    ([
      publicSansRegular,
      publicSansBold,
      newsreaderRegular,
      newsreaderBold,
      notoSansRegular,
      notoSansBold,
    ]) => [
      { data: arrayBuffer(publicSansRegular), name: "Public Sans", style: "normal", weight: 400 },
      { data: arrayBuffer(publicSansBold), name: "Public Sans", style: "normal", weight: 700 },
      { data: arrayBuffer(newsreaderRegular), name: "Newsreader", style: "normal", weight: 400 },
      { data: arrayBuffer(newsreaderBold), name: "Newsreader", style: "normal", weight: 700 },
      { data: arrayBuffer(notoSansRegular), name: "Noto Sans SC", style: "normal", weight: 400 },
      { data: arrayBuffer(notoSansBold), name: "Noto Sans SC", style: "normal", weight: 700 },
    ],
  );

  return fontsPromise;
}
