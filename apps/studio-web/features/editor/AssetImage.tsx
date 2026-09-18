import { useEffect, useState } from "react";
import { ImageOff } from "lucide-react";
import { Asset } from "../../lib/api";

export function AssetImage({asset,recipe="thumbnail",alt,className=""}:{asset:Asset;recipe?:"thumbnail"|"content-1280";alt:string;className?:string}){
  const original=`/media/${asset.id}`;
  const preferred=`${original}/${recipe}`;
  const [src,setSrc]=useState(preferred),[failed,setFailed]=useState(false);

  useEffect(()=>{setSrc(preferred);setFailed(false)},[preferred]);

  return <span className={`asset-image ${className}`.trim()}>
    {failed?<span className="asset-image-fallback" role="img" aria-label={`${alt} 暂无预览`}><ImageOff aria-hidden/><small>暂无预览</small></span>:<img src={src} alt={alt} loading="lazy" onError={()=>{if(src!==original)setSrc(original);else setFailed(true)}}/>}
  </span>;
}
