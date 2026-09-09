import type {MetadataRoute} from "next";
import {getPublicSettings,publicBase} from "../lib/content";
export default async function robots():Promise<MetadataRoute.Robots>{const settings=await getPublicSettings(),base=publicBase(settings);return{rules:{userAgent:"*",allow:"/",disallow:["/preview/"]},sitemap:new URL("/sitemap.xml",base).toString()}}
