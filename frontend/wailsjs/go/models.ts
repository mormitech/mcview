export namespace main {
	
	export class ImageInfo {
	    name: string;
	    path: string;
	    dir: string;
	    url: string;
	    index: number;
	    count: number;
	    state: number;
	
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

}

