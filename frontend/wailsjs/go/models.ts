export namespace core {
	
	export class DetectionResult {
	    detected: boolean;
	    confidence: number;
	    version?: string;
	    markers?: string[];
	
	    static createFrom(source: any = {}) {
	        return new DetectionResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.detected = source["detected"];
	        this.confidence = source["confidence"];
	        this.version = source["version"];
	        this.markers = source["markers"];
	    }
	}
	export class EndpointSource {
	    file?: string;
	    line?: number;
	
	    static createFrom(source: any = {}) {
	        return new EndpointSource(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.file = source["file"];
	        this.line = source["line"];
	    }
	}
	export class Parameter {
	    name: string;
	    in: string;
	    type?: string;
	    required: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Parameter(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.in = source["in"];
	        this.type = source["type"];
	        this.required = source["required"];
	    }
	}
	export class Endpoint {
	    id: string;
	    method: string;
	    path: string;
	    name?: string;
	    handler?: string;
	    middleware?: string[];
	    parameters?: Parameter[];
	    tags?: string[];
	    source: EndpointSource;
	    metadata?: Record<string, string>;
	
	    static createFrom(source: any = {}) {
	        return new Endpoint(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.method = source["method"];
	        this.path = source["path"];
	        this.name = source["name"];
	        this.handler = source["handler"];
	        this.middleware = source["middleware"];
	        this.parameters = this.convertValues(source["parameters"], Parameter);
	        this.tags = source["tags"];
	        this.source = this.convertValues(source["source"], EndpointSource);
	        this.metadata = source["metadata"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	

}

export namespace workspace {
	
	export class Workspace {
	    path: string;
	    name: string;
	
	    static createFrom(source: any = {}) {
	        return new Workspace(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.name = source["name"];
	    }
	}

}

