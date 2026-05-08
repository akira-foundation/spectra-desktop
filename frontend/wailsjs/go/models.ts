export namespace app {
	
	export class APIDetection {
	    mode: string;
	    value: string;
	    count: number;
	    totalCount: number;
	    scanError?: string;
	    loginRoute?: string;
	    logoutRoute?: string;
	
	    static createFrom(source: any = {}) {
	        return new APIDetection(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.mode = source["mode"];
	        this.value = source["value"];
	        this.count = source["count"];
	        this.totalCount = source["totalCount"];
	        this.scanError = source["scanError"];
	        this.loginRoute = source["loginRoute"];
	        this.logoutRoute = source["logoutRoute"];
	    }
	}
	export class EnvironmentDTO {
	    id: string;
	    projectID: string;
	    name: string;
	    vars: Record<string, string>;
	    sortOrder: number;
	
	    static createFrom(source: any = {}) {
	        return new EnvironmentDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.projectID = source["projectID"];
	        this.name = source["name"];
	        this.vars = source["vars"];
	        this.sortOrder = source["sortOrder"];
	    }
	}
	export class ExecuteRequestInput {
	    projectID: string;
	    endpointID?: string;
	    method: string;
	    path: string;
	    headers?: Record<string, string>;
	    body?: string;
	    baseUrl?: string;
	    timeoutMs?: number;
	    skipAuth?: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ExecuteRequestInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.projectID = source["projectID"];
	        this.endpointID = source["endpointID"];
	        this.method = source["method"];
	        this.path = source["path"];
	        this.headers = source["headers"];
	        this.body = source["body"];
	        this.baseUrl = source["baseUrl"];
	        this.timeoutMs = source["timeoutMs"];
	        this.skipAuth = source["skipAuth"];
	    }
	}
	export class HistoryEntryDetail {
	    id: string;
	    endpointID?: string;
	    method: string;
	    url: string;
	    responseStatus: number;
	    durationMs: number;
	    sizeBytes: number;
	    error?: string;
	    // Go type: time
	    createdAt: any;
	    requestHeaders: string;
	    requestBody: string;
	    responseHeaders: string;
	    responseBody: string;
	
	    static createFrom(source: any = {}) {
	        return new HistoryEntryDetail(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.endpointID = source["endpointID"];
	        this.method = source["method"];
	        this.url = source["url"];
	        this.responseStatus = source["responseStatus"];
	        this.durationMs = source["durationMs"];
	        this.sizeBytes = source["sizeBytes"];
	        this.error = source["error"];
	        this.createdAt = this.convertValues(source["createdAt"], null);
	        this.requestHeaders = source["requestHeaders"];
	        this.requestBody = source["requestBody"];
	        this.responseHeaders = source["responseHeaders"];
	        this.responseBody = source["responseBody"];
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
	export class HistoryListItem {
	    id: string;
	    endpointID?: string;
	    method: string;
	    url: string;
	    responseStatus: number;
	    durationMs: number;
	    sizeBytes: number;
	    error?: string;
	    // Go type: time
	    createdAt: any;
	
	    static createFrom(source: any = {}) {
	        return new HistoryListItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.endpointID = source["endpointID"];
	        this.method = source["method"];
	        this.url = source["url"];
	        this.responseStatus = source["responseStatus"];
	        this.durationMs = source["durationMs"];
	        this.sizeBytes = source["sizeBytes"];
	        this.error = source["error"];
	        this.createdAt = this.convertValues(source["createdAt"], null);
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
	export class ProjectAuthState {
	    projectID: string;
	    scheme: string;
	    hasToken: boolean;
	    tokenPreview?: string;
	    tokenPath?: string;
	    user?: core.AuthUser;
	    hasCookies: boolean;
	    // Go type: time
	    expiresAt?: any;
	    capturedFromEndpoint?: string;
	    // Go type: time
	    capturedAt: any;
	
	    static createFrom(source: any = {}) {
	        return new ProjectAuthState(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.projectID = source["projectID"];
	        this.scheme = source["scheme"];
	        this.hasToken = source["hasToken"];
	        this.tokenPreview = source["tokenPreview"];
	        this.tokenPath = source["tokenPath"];
	        this.user = this.convertValues(source["user"], core.AuthUser);
	        this.hasCookies = source["hasCookies"];
	        this.expiresAt = this.convertValues(source["expiresAt"], null);
	        this.capturedFromEndpoint = source["capturedFromEndpoint"];
	        this.capturedAt = this.convertValues(source["capturedAt"], null);
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
	export class ProjectInfo {
	    path: string;
	    name: string;
	    framework: string;
	    frameworkVersion: string;
	    detection: core.DetectionResult;
	    apiDetection: APIDetection;
	    defaultBaseUrl: string;
	    defaultPorts?: number[];
	
	    static createFrom(source: any = {}) {
	        return new ProjectInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.name = source["name"];
	        this.framework = source["framework"];
	        this.frameworkVersion = source["frameworkVersion"];
	        this.detection = this.convertValues(source["detection"], core.DetectionResult);
	        this.apiDetection = this.convertValues(source["apiDetection"], APIDetection);
	        this.defaultBaseUrl = source["defaultBaseUrl"];
	        this.defaultPorts = source["defaultPorts"];
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
	export class SaveEnvironmentInput {
	    id?: string;
	    projectID: string;
	    name: string;
	    vars?: Record<string, string>;
	    sortOrder?: number;
	
	    static createFrom(source: any = {}) {
	        return new SaveEnvironmentInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.projectID = source["projectID"];
	        this.name = source["name"];
	        this.vars = source["vars"];
	        this.sortOrder = source["sortOrder"];
	    }
	}
	export class SetProjectAuthInput {
	    projectID: string;
	    scheme: string;
	    token: string;
	
	    static createFrom(source: any = {}) {
	        return new SetProjectAuthInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.projectID = source["projectID"];
	        this.scheme = source["scheme"];
	        this.token = source["token"];
	    }
	}

}

export namespace core {
	
	export class AuthUser {
	    id?: string;
	    name?: string;
	    username?: string;
	    email?: string;
	    role?: string;
	    raw?: string;
	
	    static createFrom(source: any = {}) {
	        return new AuthUser(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.username = source["username"];
	        this.email = source["email"];
	        this.role = source["role"];
	        this.raw = source["raw"];
	    }
	}
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
	    framework?: string;
	    confidence?: number;
	    requestSchema?: string;
	    authRole?: string;
	    authHint?: string;
	    authRoleOverride?: string;
	    tokenPathOverride?: string;
	
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
	        this.framework = source["framework"];
	        this.confidence = source["confidence"];
	        this.requestSchema = source["requestSchema"];
	        this.authRole = source["authRole"];
	        this.authHint = source["authHint"];
	        this.authRoleOverride = source["authRoleOverride"];
	        this.tokenPathOverride = source["tokenPathOverride"];
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
	
	
	export class StatCard {
	    key: string;
	    kind: string;
	    label: string;
	    value: number;
	    hint?: string;
	
	    static createFrom(source: any = {}) {
	        return new StatCard(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.key = source["key"];
	        this.kind = source["kind"];
	        this.label = source["label"];
	        this.value = source["value"];
	        this.hint = source["hint"];
	    }
	}
	export class StatsReport {
	    cards: StatCard[];
	
	    static createFrom(source: any = {}) {
	        return new StatsReport(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.cards = this.convertValues(source["cards"], StatCard);
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

export namespace domain {
	
	export class Project {
	    id: string;
	    name: string;
	    path: string;
	    framework: string;
	    frameworkVersion: string;
	    status: string;
	    apiFilterMode: string;
	    apiFilterValue: string;
	    baseUrl: string;
	    loginEndpointId?: string;
	    loginTokenPath?: string;
	    logoutEndpointId?: string;
	    activeEnvironmentId?: string;
	    // Go type: time
	    createdAt: any;
	    // Go type: time
	    updatedAt: any;
	    // Go type: time
	    lastSyncedAt?: any;
	
	    static createFrom(source: any = {}) {
	        return new Project(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.path = source["path"];
	        this.framework = source["framework"];
	        this.frameworkVersion = source["frameworkVersion"];
	        this.status = source["status"];
	        this.apiFilterMode = source["apiFilterMode"];
	        this.apiFilterValue = source["apiFilterValue"];
	        this.baseUrl = source["baseUrl"];
	        this.loginEndpointId = source["loginEndpointId"];
	        this.loginTokenPath = source["loginTokenPath"];
	        this.logoutEndpointId = source["logoutEndpointId"];
	        this.activeEnvironmentId = source["activeEnvironmentId"];
	        this.createdAt = this.convertValues(source["createdAt"], null);
	        this.updatedAt = this.convertValues(source["updatedAt"], null);
	        this.lastSyncedAt = this.convertValues(source["lastSyncedAt"], null);
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
	export class ProjectInput {
	    id: string;
	    name: string;
	    path: string;
	    framework: string;
	    frameworkVersion: string;
	    apiFilterMode: string;
	    apiFilterValue: string;
	    baseUrl: string;
	
	    static createFrom(source: any = {}) {
	        return new ProjectInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.path = source["path"];
	        this.framework = source["framework"];
	        this.frameworkVersion = source["frameworkVersion"];
	        this.apiFilterMode = source["apiFilterMode"];
	        this.apiFilterValue = source["apiFilterValue"];
	        this.baseUrl = source["baseUrl"];
	    }
	}
	export class ProjectStats {
	    routes: number;
	    models: number;
	    middleware: number;
	    controllers: number;
	    errors: number;
	    // Go type: time
	    lastScannedAt?: any;
	
	    static createFrom(source: any = {}) {
	        return new ProjectStats(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.routes = source["routes"];
	        this.models = source["models"];
	        this.middleware = source["middleware"];
	        this.controllers = source["controllers"];
	        this.errors = source["errors"];
	        this.lastScannedAt = this.convertValues(source["lastScannedAt"], null);
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

export namespace httpclient {
	
	export class Response {
	    status: number;
	    statusText: string;
	    headers?: Record<string, Array<string>>;
	    body?: string;
	    durationMs: number;
	    sizeBytes: number;
	
	    static createFrom(source: any = {}) {
	        return new Response(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.status = source["status"];
	        this.statusText = source["statusText"];
	        this.headers = source["headers"];
	        this.body = source["body"];
	        this.durationMs = source["durationMs"];
	        this.sizeBytes = source["sizeBytes"];
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

