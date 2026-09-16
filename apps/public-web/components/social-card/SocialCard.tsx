import { BrandMark } from "../BrandMark";

export type SocialCardProps = {
  accent: string;
  coverData?: string;
  date?: string;
  description: string;
  domain: string;
  kind: "article" | "site";
  locale: string;
  siteName: string;
  title: string;
  type?: string;
};

function shortened(value: string, limit: number) {
  const characters = Array.from(value.trim());
  return characters.length > limit
    ? `${characters.slice(0, limit - 1).join("")}…`
    : characters.join("");
}

function Observatory({ accent }: { accent: string }) {
  return (
    <div
      style={{
        alignItems: "center",
        display: "flex",
        height: "100%",
        justifyContent: "center",
        position: "relative",
        width: "100%",
      }}
    >
      <div style={{ border: "1px solid rgba(23,20,22,.16)", borderRadius: 999, height: 430, position: "absolute", transform: "rotate(-14deg)", width: 430 }} />
      <div style={{ border: "1px solid rgba(23,20,22,.12)", borderRadius: "50%", height: 212, position: "absolute", transform: "rotate(18deg)", width: 440 }} />
      <div style={{ border: `3px solid ${accent}`, borderRadius: "50%", height: 96, position: "absolute", transform: "rotate(-9deg)", width: 330 }} />
      <div style={{ background: "#171416", border: `7px solid ${accent}`, borderRadius: 999, boxShadow: "0 0 0 13px rgba(23,20,22,.08)", display: "flex", height: 168, width: 168 }} />
      <div style={{ background: accent, borderRadius: 999, display: "flex", height: 13, position: "absolute", right: 72, top: 154, width: 13 }} />
      <div style={{ bottom: 64, color: "rgba(23,20,22,.5)", display: "flex", fontSize: 15, letterSpacing: ".2em", position: "absolute", right: 55 }}>BH–01 / TRUTH OBSERVATORY</div>
    </div>
  );
}

function Visual({ accent, coverData }: { accent: string; coverData?: string }) {
  return (
    <div
      style={{
        alignItems: "center",
        backgroundColor: "#ece7de",
        backgroundPosition: "center",
        backgroundSize: "cover",
        borderLeft: "1px solid rgba(23,20,22,.18)",
        display: "flex",
        height: "100%",
        justifyContent: "center",
        overflow: "hidden",
        position: "absolute",
        right: 0,
        top: 0,
        width: 490,
        ...(coverData ? { backgroundImage: `url(${coverData})` } : {}),
      }}
    >
      {coverData ? (
        <>
          <div style={{ background: "linear-gradient(90deg,rgba(23,20,22,.22),transparent 42%)", display: "flex", height: "100%", left: 0, position: "absolute", top: 0, width: "100%" }} />
          <div style={{ border: "1px solid rgba(244,240,232,.72)", borderRadius: "50%", display: "flex", height: 330, position: "absolute", right: -102, top: 145, width: 330 }} />
          <div style={{ background: accent, borderRadius: 999, display: "flex", height: 16, position: "absolute", right: 38, top: 108, width: 16 }} />
        </>
      ) : <Observatory accent={accent} />}
    </div>
  );
}

export function SocialCard(props: SocialCardProps) {
  const zh = props.locale === "zh-CN" || props.locale === "zh";
  const titleLimit = props.kind === "article" ? (zh ? 42 : 76) : 54;
  const descriptionLimit = zh ? 72 : 130;
  const date = props.date
    ? new Intl.DateTimeFormat(zh ? "zh-CN" : "en", { dateStyle: "medium" }).format(new Date(props.date))
    : "";

  return (
    <div
      style={{
        background: "#f4f0e8",
        color: "#171416",
        display: "flex",
        height: "100%",
        overflow: "hidden",
        position: "relative",
        width: "100%",
      }}
    >
      <div style={{ background: props.accent, display: "flex", height: 9, left: 0, position: "absolute", top: 0, width: 710 }} />
      <div style={{ display: "flex", flexDirection: "column", height: "100%", padding: "52px 58px 48px", width: 710 }}>
        <div style={{ alignItems: "center", display: "flex", justifyContent: "space-between", width: "100%" }}>
          <div style={{ alignItems: "center", display: "flex" }}>
            <BrandMark size={45} />
            <div style={{ display: "flex", fontSize: 20, fontWeight: 800, letterSpacing: "-.02em", marginLeft: 13, textTransform: "uppercase" }}>{shortened(props.siteName || "Field Notes", 32)}</div>
          </div>
          <div style={{ color: props.accent, display: "flex", fontSize: 14, fontWeight: 700, letterSpacing: ".18em" }}>{props.kind === "article" ? "FIELD NOTE / 02" : "OBSERVATORY / 01"}</div>
        </div>

        <div style={{ display: "flex", flexDirection: "column", flexGrow: 1, justifyContent: "center", paddingTop: 24 }}>
          {props.kind === "article" && <div style={{ color: props.accent, display: "flex", fontSize: 16, fontWeight: 800, letterSpacing: ".15em", marginBottom: 22, textTransform: "uppercase" }}>{props.type || "ARTICLE"}{date ? `  ·  ${date}` : ""}</div>}
          <div style={{ display: "flex", fontSize: props.kind === "article" ? 62 : 68, fontWeight: 650, letterSpacing: "-.055em", lineHeight: 1.04, maxHeight: 210, overflow: "hidden" }}>{shortened(props.title, titleLimit)}</div>
          <div style={{ color: "#686168", display: "flex", fontSize: 24, lineHeight: 1.48, marginTop: 24, maxHeight: 108, overflow: "hidden" }}>{shortened(props.description, descriptionLimit)}</div>
        </div>

        <div style={{ alignItems: "center", borderTop: "1px solid rgba(23,20,22,.24)", display: "flex", fontSize: 15, justifyContent: "space-between", letterSpacing: ".12em", paddingTop: 17, textTransform: "uppercase", width: "100%" }}>
          <span>{props.domain}</span>
          <span style={{ color: "#686168" }}>{zh ? "持续记录 · 独立思考" : "CAPTURE · CONNECT · PUBLISH"}</span>
        </div>
      </div>
      <Visual accent={props.accent} coverData={props.coverData} />
    </div>
  );
}
