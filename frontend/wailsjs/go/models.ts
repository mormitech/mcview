export namespace main {
	
	export class GifMeta {
	    frames: number;
	    delays: number[];
	
	    static createFrom(source: any = {}) {
	        return new GifMeta(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.frames = source["frames"];
	        this.delays = source["delays"];
	    }
	}
	export class ImageInfo {
	    name: string;
	    path: string;
	    dir: string;
	    url: string;
	    index: number;
	    count: number;
	    state: number;
	    modified: boolean;
	    canAlpha: boolean;
	    note?: string;
	
	    static createFrom(source: any = {}) {
	        return new ImageInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.path = source["path"];
	        this.dir = source["dir"];
	        this.url = source["url"];
	        this.index = source["index"];
	        this.count = source["count"];
	        this.state = source["state"];
	        this.modified = source["modified"];
	        this.canAlpha = source["canAlpha"];
	        this.note = source["note"];
	    }
	}
	export class PastedLayer {
	    png: string;
	    x: number;
	    y: number;
	    w: number;
	    h: number;
	
	    static createFrom(source: any = {}) {
	        return new PastedLayer(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.png = source["png"];
	        this.x = source["x"];
	        this.y = source["y"];
	        this.w = source["w"];
	        this.h = source["h"];
	    }
	}
	export class Pixel {
	    ready: boolean;
	    x: number;
	    y: number;
	    r: number;
	    g: number;
	    b: number;
	    a: number;
	
	    static createFrom(source: any = {}) {
	        return new Pixel(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ready = source["ready"];
	        this.x = source["x"];
	        this.y = source["y"];
	        this.r = source["r"];
	        this.g = source["g"];
	        this.b = source["b"];
	        this.a = source["a"];
	    }
	}
	export class RecentFile {
	    name: string;
	    dir: string;
	    path: string;
	
	    static createFrom(source: any = {}) {
	        return new RecentFile(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.dir = source["dir"];
	        this.path = source["path"];
	    }
	}

}

