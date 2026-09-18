import { useMemo, useState } from "react";
import { Check, Image as ImageIcon, Search } from "lucide-react";
import { Asset } from "../../lib/api";
import { AssetImage } from "./AssetImage";

function assetSize(asset:Asset){return `${Math.ceil(asset.size/1024)} KB${asset.width&&asset.height?` · ${asset.width} × ${asset.height}`:""}`}

export function ImageAssetBrowser({items,selected,onSelect}:{items:Asset[];selected:string;onSelect:(id:string)=>void}){
  const [query,setQuery]=useState("");
  const filtered=useMemo(()=>{const keyword=query.trim().toLowerCase();return keyword?items.filter(item=>item.filename.toLowerCase().includes(keyword)):items},[items,query]);
  const selectedAsset=items.find(item=>item.id===selected);

  return <div className="image-asset-browser">
    <label className="image-asset-search"><Search aria-hidden/><span className="sr-only">搜索图片</span><input value={query} onChange={event=>setQuery(event.target.value)} placeholder="搜索图片文件名…"/></label>
    <div className="asset-picker-browser">
      <div className="image-asset-grid" role="radiogroup" aria-label="图片资产">
        {filtered.map(item=><button type="button" role="radio" aria-checked={selected===item.id} key={item.id} className={selected===item.id?"image-asset-card selected":"image-asset-card"} onClick={()=>onSelect(item.id)}>
          <AssetImage asset={item} alt={item.filename}/>
          <span className="image-asset-card-copy"><strong title={item.filename}>{item.filename}</strong><small>{assetSize(item)}</small></span>
          {selected===item.id&&<span className="image-asset-selected" aria-hidden><Check/></span>}
        </button>)}
        {filtered.length===0&&<div className="image-asset-empty"><ImageIcon aria-hidden/><strong>{items.length?"没有匹配的图片":"资产库中还没有图片"}</strong><small>{items.length?"尝试其他文件名":"请先上传图片后再选择"}</small></div>}
      </div>
      {selectedAsset?<aside className="asset-picker-preview" aria-live="polite"><AssetImage asset={selectedAsset} recipe="content-1280" alt={selectedAsset.filename}/><span className="selected-kicker"><Check aria-hidden/>当前选中</span><strong title={selectedAsset.filename}>{selectedAsset.filename}</strong><small>{selectedAsset.mime} · {assetSize(selectedAsset)}</small></aside>:<div className="asset-picker-preview empty"><ImageIcon aria-hidden/><strong>尚未选择图片</strong><span>点击左侧图片查看大图</span></div>}
    </div>
  </div>;
}
