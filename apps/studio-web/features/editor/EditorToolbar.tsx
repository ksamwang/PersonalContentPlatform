import type { Editor } from "@tiptap/react";
import type { ReactNode } from "react";
import {
  Bold, Braces, Code2, Heading1, Heading2, Heading3, ImagePlus, Italic,
  Link2, List, ListChecks, ListOrdered, Maximize2, Minimize2, Minus,
  Quote, Redo2, Strikethrough, Table2, Underline as UnderlineIcon, Undo2, Unlink,
} from "lucide-react";

type ToolProps = { label:string; active?:boolean; disabled?:boolean; onClick:()=>void; children:ReactNode };
function Tool({label,active,disabled,onClick,children}:ToolProps){
  return <button type="button" aria-label={label} title={label} aria-pressed={active||undefined} disabled={disabled} onClick={onClick}>{children}</button>;
}

export function EditorToolbar({editor,focusMode,onFocusMode}:{editor:Editor;focusMode:boolean;onFocusMode:()=>void}){
  function editLink(){
    const previous=editor.getAttributes("link").href as string|undefined;
    const href=window.prompt("输入链接地址",previous||"https://");
    if(href===null)return;
    if(!href.trim()){editor.chain().focus().extendMarkRange("link").unsetLink().run();return}
    editor.chain().focus().extendMarkRange("link").setLink({href:href.trim()}).run();
  }
  function insertImage(){
    const src=window.prompt("输入图片地址");
    if(!src?.trim())return;
    const alt=window.prompt("输入图片替代文字（用于无障碍阅读）")?.trim()||"";
    editor.chain().focus().setImage({src:src.trim(),alt}).run();
  }
  return <div className="toolbar" role="toolbar" aria-label="编辑工具栏">
    <div className="toolbar-group" aria-label="历史操作">
      <Tool label="撤销 Ctrl+Z" disabled={!editor.can().undo()} onClick={()=>editor.chain().focus().undo().run()}><Undo2/></Tool>
      <Tool label="重做 Ctrl+Shift+Z" disabled={!editor.can().redo()} onClick={()=>editor.chain().focus().redo().run()}><Redo2/></Tool>
    </div>
    <div className="toolbar-group" aria-label="文本样式">
      <Tool label="一级标题" active={editor.isActive("heading",{level:1})} onClick={()=>editor.chain().focus().toggleHeading({level:1}).run()}><Heading1/></Tool>
      <Tool label="二级标题" active={editor.isActive("heading",{level:2})} onClick={()=>editor.chain().focus().toggleHeading({level:2}).run()}><Heading2/></Tool>
      <Tool label="三级标题" active={editor.isActive("heading",{level:3})} onClick={()=>editor.chain().focus().toggleHeading({level:3}).run()}><Heading3/></Tool>
      <Tool label="粗体 Ctrl+B" active={editor.isActive("bold")} onClick={()=>editor.chain().focus().toggleBold().run()}><Bold/></Tool>
      <Tool label="斜体 Ctrl+I" active={editor.isActive("italic")} onClick={()=>editor.chain().focus().toggleItalic().run()}><Italic/></Tool>
      <Tool label="下划线 Ctrl+U" active={editor.isActive("underline")} onClick={()=>editor.chain().focus().toggleUnderline().run()}><UnderlineIcon/></Tool>
      <Tool label="删除线" active={editor.isActive("strike")} onClick={()=>editor.chain().focus().toggleStrike().run()}><Strikethrough/></Tool>
      <Tool label="行内代码" active={editor.isActive("code")} onClick={()=>editor.chain().focus().toggleCode().run()}><Code2/></Tool>
    </div>
    <div className="toolbar-group" aria-label="段落结构">
      <Tool label="无序列表" active={editor.isActive("bulletList")} onClick={()=>editor.chain().focus().toggleBulletList().run()}><List/></Tool>
      <Tool label="有序列表" active={editor.isActive("orderedList")} onClick={()=>editor.chain().focus().toggleOrderedList().run()}><ListOrdered/></Tool>
      <Tool label="任务列表" active={editor.isActive("taskList")} onClick={()=>editor.chain().focus().toggleTaskList().run()}><ListChecks/></Tool>
      <Tool label="引用" active={editor.isActive("blockquote")} onClick={()=>editor.chain().focus().toggleBlockquote().run()}><Quote/></Tool>
      <Tool label="代码块" active={editor.isActive("codeBlock")} onClick={()=>editor.chain().focus().toggleCodeBlock().run()}><Braces/></Tool>
      <Tool label="分隔线" onClick={()=>editor.chain().focus().setHorizontalRule().run()}><Minus/></Tool>
    </div>
    <div className="toolbar-group" aria-label="插入内容">
      <Tool label="添加或编辑链接" active={editor.isActive("link")} onClick={editLink}><Link2/></Tool>
      <Tool label="移除链接" disabled={!editor.isActive("link")} onClick={()=>editor.chain().focus().unsetLink().run()}><Unlink/></Tool>
      <Tool label="插入图片" onClick={insertImage}><ImagePlus/></Tool>
      <Tool label="插入 3×3 表格" onClick={()=>editor.chain().focus().insertTable({rows:3,cols:3,withHeaderRow:true}).run()}><Table2/></Tool>
    </div>
    <div className="toolbar-group toolbar-end">
      <Tool label={focusMode?"退出专注模式":"进入专注模式"} active={focusMode} onClick={onFocusMode}>{focusMode?<Minimize2/>:<Maximize2/>}</Tool>
    </div>
  </div>;
}
