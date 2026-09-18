import { useState } from "react";
import { X } from "lucide-react";
import { Asset } from "../../lib/api";
import { ImageAssetBrowser } from "./ImageAssetBrowser";

export function CoverAssetPickerDialog({items,selected,onSelect,onClose}:{items:Asset[];selected:string;onSelect:(id:string)=>void;onClose:()=>void}){
  const [nextSelected,setNextSelected]=useState(selected);
  return <div className="modal-backdrop" role="presentation" onMouseDown={event=>{if(event.target===event.currentTarget)onClose()}}><section className="modal asset-picker" role="dialog" aria-modal="true" aria-labelledby="cover-picker-title">
    <button className="icon-button close" aria-label="关闭封面选择" onClick={onClose}><X/></button>
    <h2 id="cover-picker-title">选择文章封面</h2>
    <p>从图片库中查看并选择封面，确认前不会替换当前图片。</p>
    <ImageAssetBrowser items={items} selected={nextSelected} onSelect={setNextSelected}/>
    <div className="asset-picker-actions"><button type="button" className="secondary" onClick={onClose}>取消</button><button type="button" className="primary" disabled={!nextSelected} onClick={()=>{onSelect(nextSelected);onClose()}}>使用这张图片</button></div>
  </section></div>;
}
