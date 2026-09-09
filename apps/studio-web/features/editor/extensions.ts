import Image from "@tiptap/extension-image";
import Link from "@tiptap/extension-link";
import { Table, TableCell, TableHeader, TableRow } from "@tiptap/extension-table";
import TaskItem from "@tiptap/extension-task-item";
import TaskList from "@tiptap/extension-task-list";
import Underline from "@tiptap/extension-underline";
import StarterKit from "@tiptap/starter-kit";

export const editorExtensions = [
  StarterKit.configure({ heading: { levels: [1, 2, 3] } }),
  Underline,
  Link.configure({ autolink: true, openOnClick: false, defaultProtocol: "https" }),
  TaskList,
  TaskItem.configure({ nested: true }),
  Image.configure({ allowBase64: false }),
  Table.configure({ resizable: true }),
  TableRow,
  TableHeader,
  TableCell,
];
