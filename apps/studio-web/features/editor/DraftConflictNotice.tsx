import { Copy, RefreshCw, Upload } from "lucide-react";

export function DraftConflictNotice({onReload,onCopy,onOverwrite}:{onReload:()=>void;onCopy:()=>void;onOverwrite:()=>void}){
  return <section className="draft-conflict" role="alert"><div><strong>检测到其他位置保存的新版本</strong><span>请选择重新加载，或保留当前编辑内容并覆盖服务器草稿。</span></div><div><button className="secondary compact" onClick={onCopy}><Copy aria-hidden/>复制本地内容</button><button className="secondary compact" onClick={onReload}><RefreshCw aria-hidden/>加载服务器版本</button><button className="primary compact" onClick={onOverwrite}><Upload aria-hidden/>使用本地内容</button></div></section>;
}
