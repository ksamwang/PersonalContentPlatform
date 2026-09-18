import Link from "next/link";
import { ArrowUpRight } from "lucide-react";
import { PublishedPage } from "../lib/content";
import { normalizeTags } from "../lib/tags";

export function ContentCards({items,locale}:{items:PublishedPage[];locale:string}) {
  return <div className="publication-index">{items.map(item=>{
    const cover=item.metadata?.cover_asset_id;
    const tags=normalizeTags(item.metadata?.tags).slice(0,4);
    return <Link className={cover?"":"without-cover"} key={`${item.content_id||item.slug}-${item.locale}`} href={`/${locale}/${item.type}/${item.slug}`}>
      {cover?<picture><source media="(max-width: 640px)" srcSet={`/media/${cover}/thumbnail`}/><img loading="lazy" src={`/media/${cover}/content-1280`} alt=""/></picture>:null}
      <div className="publication-copy">
        <div className="publication-meta"><span>{item.type}</span><time dateTime={item.published_at}>{new Intl.DateTimeFormat(locale,{year:"numeric",month:"short",day:"numeric"}).format(new Date(item.published_at))}</time></div>
        <strong>{item.title}</strong>
        {item.summary&&<p>{item.summary}</p>}
        {tags.length>0&&<small>{tags.join(" · ")}</small>}
      </div>
      <ArrowUpRight aria-hidden />
    </Link>
  })}</div>;
}
