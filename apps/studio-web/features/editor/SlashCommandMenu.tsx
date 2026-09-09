import type { Editor } from "@tiptap/react";
import { Braces, Heading2, ImagePlus, List, ListChecks, Quote, Table2 } from "lucide-react";

export function SlashCommandMenu({editor,open,onClose}:{editor:Editor;open:boolean;onClose:()=>void}){
  if(!open)return null;
  const run=(command:(editor:Editor)=>void)=>{const from=editor.state.selection.from;editor.chain().focus().deleteRange({from:Math.max(0,from-1),to:from}).run();command(editor);onClose()};
  const items=[
    ["二级标题","章节标题",Heading2,(e:Editor)=>e.chain().focus().toggleHeading({level:2}).run()],
    ["无序列表","创建项目列表",List,(e:Editor)=>e.chain().focus().toggleBulletList().run()],
    ["任务列表","创建待办事项",ListChecks,(e:Editor)=>e.chain().focus().toggleTaskList().run()],
    ["引用","突出显示引用内容",Quote,(e:Editor)=>e.chain().focus().toggleBlockquote().run()],
    ["代码块","插入代码段",Braces,(e:Editor)=>e.chain().focus().toggleCodeBlock().run()],
    ["表格","插入 3×3 表格",Table2,(e:Editor)=>e.chain().focus().insertTable({rows:3,cols:3,withHeaderRow:true}).run()],
    ["图片","通过地址插入图片",ImagePlus,(e:Editor)=>{const src=window.prompt("输入图片地址");if(src?.trim())e.chain().focus().setImage({src:src.trim(),alt:""}).run()}],
  ] as const;
  return <div className="slash-menu" role="menu" aria-label="插入内容">{items.map(([name,description,Icon,command])=><button type="button" role="menuitem" key={name} onClick={()=>run(command)}><Icon aria-hidden/><span><strong>{name}</strong><small>{description}</small></span></button>)}</div>;
}
