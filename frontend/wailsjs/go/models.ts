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
	export class CapturedValueDTO {
	    name: string;
	    value: string;
	    endpointKey?: string;
	    capturedAt?: number;
	
	    static createFrom(source: any = {}) {
	        return new CapturedValueDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.value = source["value"];
	        this.endpointKey = source["endpointKey"];
	        this.capturedAt = source["capturedAt"];
	    }
	}
	export class CollectionItemDTO {
	    id?: string;
	    endpointID: string;
	    bodyOverride?: string;
	    headersOverride?: string;
	    skipOnFailure?: boolean;
	    iterateDataset?: boolean;
	
	    static createFrom(source: any = {}) {
	        return new CollectionItemDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.endpointID = source["endpointID"];
	        this.bodyOverride = source["bodyOverride"];
	        this.headersOverride = source["headersOverride"];
	        this.skipOnFailure = source["skipOnFailure"];
	        this.iterateDataset = source["iterateDataset"];
	    }
	}
	export class CollectionDTO {
	    id: string;
	    projectID: string;
	    name: string;
	    description?: string;
	    sortOrder: number;
	    items: CollectionItemDTO[];
	
	    static createFrom(source: any = {}) {
	        return new CollectionDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.projectID = source["projectID"];
	        this.name = source["name"];
	        this.description = source["description"];
	        this.sortOrder = source["sortOrder"];
	        this.items = this.convertValues(source["items"], CollectionItemDTO);
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
	
	export class TestResultDTO {
	    id?: string;
	    name: string;
	    kind: string;
	    pass: boolean;
	    message?: string;
	
	    static createFrom(source: any = {}) {
	        return new TestResultDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.kind = source["kind"];
	        this.pass = source["pass"];
	        this.message = source["message"];
	    }
	}
	export class CollectionRunItemDTO {
	    endpointID: string;
	    method: string;
	    path: string;
	    status: number;
	    durationMs: number;
	    pass: boolean;
	    skipped?: boolean;
	    error?: string;
	    testResults?: TestResultDTO[];
	
	    static createFrom(source: any = {}) {
	        return new CollectionRunItemDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.endpointID = source["endpointID"];
	        this.method = source["method"];
	        this.path = source["path"];
	        this.status = source["status"];
	        this.durationMs = source["durationMs"];
	        this.pass = source["pass"];
	        this.skipped = source["skipped"];
	        this.error = source["error"];
	        this.testResults = this.convertValues(source["testResults"], TestResultDTO);
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
	export class CollectionRunDTO {
	    collectionID: string;
	    startedAt: number;
	    durationMs: number;
	    passCount: number;
	    failCount: number;
	    skipCount: number;
	    items: CollectionRunItemDTO[];
	
	    static createFrom(source: any = {}) {
	        return new CollectionRunDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.collectionID = source["collectionID"];
	        this.startedAt = source["startedAt"];
	        this.durationMs = source["durationMs"];
	        this.passCount = source["passCount"];
	        this.failCount = source["failCount"];
	        this.skipCount = source["skipCount"];
	        this.items = this.convertValues(source["items"], CollectionRunItemDTO);
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
	
	export class CurlImportDTO {
	    method: string;
	    url: string;
	    baseURL: string;
	    path: string;
	    headers: Record<string, string>;
	    body?: string;
	    query?: Record<string, string>;
	
	    static createFrom(source: any = {}) {
	        return new CurlImportDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.method = source["method"];
	        this.url = source["url"];
	        this.baseURL = source["baseURL"];
	        this.path = source["path"];
	        this.headers = source["headers"];
	        this.body = source["body"];
	        this.query = source["query"];
	    }
	}
	export class EndpointMetricDTO {
	    endpointID: string;
	    method: string;
	    path: string;
	    count: number;
	    errors: number;
	    avgMs: number;
	    errorRate: number;
	
	    static createFrom(source: any = {}) {
	        return new EndpointMetricDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.endpointID = source["endpointID"];
	        this.method = source["method"];
	        this.path = source["path"];
	        this.count = source["count"];
	        this.errors = source["errors"];
	        this.avgMs = source["avgMs"];
	        this.errorRate = source["errorRate"];
	    }
	}
	export class VolumePoint {
	    day: string;
	    count: number;
	
	    static createFrom(source: any = {}) {
	        return new VolumePoint(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.day = source["day"];
	        this.count = source["count"];
	    }
	}
	export class LatencyDTO {
	    count: number;
	    avg: number;
	    min: number;
	    max: number;
	    p50: number;
	    p95: number;
	    p99: number;
	
	    static createFrom(source: any = {}) {
	        return new LatencyDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.count = source["count"];
	        this.avg = source["avg"];
	        this.min = source["min"];
	        this.max = source["max"];
	        this.p50 = source["p50"];
	        this.p95 = source["p95"];
	        this.p99 = source["p99"];
	    }
	}
	export class StatusBucketDTO {
	    bucket: string;
	    count: number;
	
	    static createFrom(source: any = {}) {
	        return new StatusBucketDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.bucket = source["bucket"];
	        this.count = source["count"];
	    }
	}
	export class DashboardMetrics {
	    statusBuckets: StatusBucketDTO[];
	    latency: LatencyDTO;
	    volume: VolumePoint[];
	    totalRuns: number;
	    errorRate: number;
	    topSlow: EndpointMetricDTO[];
	    topFailing: EndpointMetricDTO[];
	    topUsed: EndpointMetricDTO[];
	
	    static createFrom(source: any = {}) {
	        return new DashboardMetrics(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.statusBuckets = this.convertValues(source["statusBuckets"], StatusBucketDTO);
	        this.latency = this.convertValues(source["latency"], LatencyDTO);
	        this.volume = this.convertValues(source["volume"], VolumePoint);
	        this.totalRuns = source["totalRuns"];
	        this.errorRate = source["errorRate"];
	        this.topSlow = this.convertValues(source["topSlow"], EndpointMetricDTO);
	        this.topFailing = this.convertValues(source["topFailing"], EndpointMetricDTO);
	        this.topUsed = this.convertValues(source["topUsed"], EndpointMetricDTO);
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
	export class DatasetRowResultDTO {
	    index: number;
	    status: number;
	    durationMs: number;
	    pass: boolean;
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new DatasetRowResultDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.index = source["index"];
	        this.status = source["status"];
	        this.durationMs = source["durationMs"];
	        this.pass = source["pass"];
	        this.error = source["error"];
	    }
	}
	export class DatasetRunDTO {
	    endpointKey: string;
	    total: number;
	    passCount: number;
	    failCount: number;
	    durationMs: number;
	    rows: DatasetRowResultDTO[];
	
	    static createFrom(source: any = {}) {
	        return new DatasetRunDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.endpointKey = source["endpointKey"];
	        this.total = source["total"];
	        this.passCount = source["passCount"];
	        this.failCount = source["failCount"];
	        this.durationMs = source["durationMs"];
	        this.rows = this.convertValues(source["rows"], DatasetRowResultDTO);
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
	export class EndpointDiscoveryDTO {
	    endpointID: string;
	    method: string;
	    path: string;
	    lastSeen?: number;
	    daysAgo?: number;
	
	    static createFrom(source: any = {}) {
	        return new EndpointDiscoveryDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.endpointID = source["endpointID"];
	        this.method = source["method"];
	        this.path = source["path"];
	        this.lastSeen = source["lastSeen"];
	        this.daysAgo = source["daysAgo"];
	    }
	}
	export class DiscoveryDTO {
	    totalEndpoints: number;
	    usedEndpoints: number;
	    coverage: number;
	    unused: EndpointDiscoveryDTO[];
	    stale: EndpointDiscoveryDTO[];
	    testedEndpoints: number;
	    testCoverage: number;
	    writeEndpoints: number;
	    readEndpoints: number;
	    authRequired: number;
	    authPublic: number;
	
	    static createFrom(source: any = {}) {
	        return new DiscoveryDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.totalEndpoints = source["totalEndpoints"];
	        this.usedEndpoints = source["usedEndpoints"];
	        this.coverage = source["coverage"];
	        this.unused = this.convertValues(source["unused"], EndpointDiscoveryDTO);
	        this.stale = this.convertValues(source["stale"], EndpointDiscoveryDTO);
	        this.testedEndpoints = source["testedEndpoints"];
	        this.testCoverage = source["testCoverage"];
	        this.writeEndpoints = source["writeEndpoints"];
	        this.readEndpoints = source["readEndpoints"];
	        this.authRequired = source["authRequired"];
	        this.authPublic = source["authPublic"];
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
	export class EndpointCaptureDTO {
	    id?: string;
	    name: string;
	    source: string;
	    path: string;
	
	    static createFrom(source: any = {}) {
	        return new EndpointCaptureDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.source = source["source"];
	        this.path = source["path"];
	    }
	}
	
	export class LatencyPointDTO {
	    day: string;
	    avgMs: number;
	    count: number;
	
	    static createFrom(source: any = {}) {
	        return new LatencyPointDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.day = source["day"];
	        this.avgMs = source["avgMs"];
	        this.count = source["count"];
	    }
	}
	export class EndpointFailureSeriesDTO {
	    endpointID: string;
	    method: string;
	    path: string;
	    failures: number;
	    points: LatencyPointDTO[];
	
	    static createFrom(source: any = {}) {
	        return new EndpointFailureSeriesDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.endpointID = source["endpointID"];
	        this.method = source["method"];
	        this.path = source["path"];
	        this.failures = source["failures"];
	        this.points = this.convertValues(source["points"], LatencyPointDTO);
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
	export class EndpointLatencySeriesDTO {
	    endpointID: string;
	    method: string;
	    path: string;
	    avgMs: number;
	    points: LatencyPointDTO[];
	
	    static createFrom(source: any = {}) {
	        return new EndpointLatencySeriesDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.endpointID = source["endpointID"];
	        this.method = source["method"];
	        this.path = source["path"];
	        this.avgMs = source["avgMs"];
	        this.points = this.convertValues(source["points"], LatencyPointDTO);
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
	
	export class EndpointTestDTO {
	    id?: string;
	    name?: string;
	    kind: string;
	    jsonPath?: string;
	    op?: string;
	    expected?: string;
	
	    static createFrom(source: any = {}) {
	        return new EndpointTestDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.kind = source["kind"];
	        this.jsonPath = source["jsonPath"];
	        this.op = source["op"];
	        this.expected = source["expected"];
	    }
	}
	export class EndpointUsageSeriesDTO {
	    endpointID: string;
	    method: string;
	    path: string;
	    total: number;
	    points: LatencyPointDTO[];
	
	    static createFrom(source: any = {}) {
	        return new EndpointUsageSeriesDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.endpointID = source["endpointID"];
	        this.method = source["method"];
	        this.path = source["path"];
	        this.total = source["total"];
	        this.points = this.convertValues(source["points"], LatencyPointDTO);
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
	export class MultipartPartDTO {
	    name: string;
	    value?: string;
	    filePath?: string;
	
	    static createFrom(source: any = {}) {
	        return new MultipartPartDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.value = source["value"];
	        this.filePath = source["filePath"];
	    }
	}
	export class ExecuteRequestInput {
	    projectID: string;
	    endpointID?: string;
	    method: string;
	    path: string;
	    headers?: Record<string, string>;
	    body?: string;
	    multipart?: MultipartPartDTO[];
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
	        this.multipart = this.convertValues(source["multipart"], MultipartPartDTO);
	        this.baseUrl = source["baseUrl"];
	        this.timeoutMs = source["timeoutMs"];
	        this.skipAuth = source["skipAuth"];
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
	export class FlakyEndpointDTO {
	    endpointID: string;
	    method: string;
	    path: string;
	    total: number;
	    successes: number;
	    failures: number;
	    flakeScore: number;
	
	    static createFrom(source: any = {}) {
	        return new FlakyEndpointDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.endpointID = source["endpointID"];
	        this.method = source["method"];
	        this.path = source["path"];
	        this.total = source["total"];
	        this.successes = source["successes"];
	        this.failures = source["failures"];
	        this.flakeScore = source["flakeScore"];
	    }
	}
	export class HAREntryDTO {
	    method: string;
	    url: string;
	    baseURL?: string;
	    path?: string;
	    headers?: Record<string, string>;
	    body?: string;
	    query?: Record<string, string>;
	    status?: number;
	    size?: number;
	    startedAt?: string;
	
	    static createFrom(source: any = {}) {
	        return new HAREntryDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.method = source["method"];
	        this.url = source["url"];
	        this.baseURL = source["baseURL"];
	        this.path = source["path"];
	        this.headers = source["headers"];
	        this.body = source["body"];
	        this.query = source["query"];
	        this.status = source["status"];
	        this.size = source["size"];
	        this.startedAt = source["startedAt"];
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
	    testResults?: TestResultDTO[];
	
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
	        this.testResults = this.convertValues(source["testResults"], TestResultDTO);
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
	export class HourlyCellDTO {
	    day: number;
	    hour: number;
	    count: number;
	
	    static createFrom(source: any = {}) {
	        return new HourlyCellDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.day = source["day"];
	        this.hour = source["hour"];
	        this.count = source["count"];
	    }
	}
	export class ImportCollectionResult {
	    collection: CollectionDTO;
	    missingEndpoints?: string[];
	
	    static createFrom(source: any = {}) {
	        return new ImportCollectionResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.collection = this.convertValues(source["collection"], CollectionDTO);
	        this.missingEndpoints = source["missingEndpoints"];
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
	export class MethodShareDTO {
	    method: string;
	    count: number;
	    percent: number;
	
	    static createFrom(source: any = {}) {
	        return new MethodShareDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.method = source["method"];
	        this.count = source["count"];
	        this.percent = source["percent"];
	    }
	}
	export class InsightsDTO {
	    latencyOverTime: EndpointLatencySeriesDTO[];
	    usageOverTime: EndpointUsageSeriesDTO[];
	    failuresOverTime: EndpointFailureSeriesDTO[];
	    hourlyHeatmap: HourlyCellDTO[];
	    flaky: FlakyEndpointDTO[];
	    methodShare: MethodShareDTO[];
	
	    static createFrom(source: any = {}) {
	        return new InsightsDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.latencyOverTime = this.convertValues(source["latencyOverTime"], EndpointLatencySeriesDTO);
	        this.usageOverTime = this.convertValues(source["usageOverTime"], EndpointUsageSeriesDTO);
	        this.failuresOverTime = this.convertValues(source["failuresOverTime"], EndpointFailureSeriesDTO);
	        this.hourlyHeatmap = this.convertValues(source["hourlyHeatmap"], HourlyCellDTO);
	        this.flaky = this.convertValues(source["flaky"], FlakyEndpointDTO);
	        this.methodShare = this.convertValues(source["methodShare"], MethodShareDTO);
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
	export class RegenerateFieldInput {
	    name: string;
	    type: string;
	    rules?: string[];
	
	    static createFrom(source: any = {}) {
	        return new RegenerateFieldInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.type = source["type"];
	        this.rules = source["rules"];
	    }
	}
	export class RegenerateBodyInput {
	    projectID?: string;
	    body: string;
	    fields?: RegenerateFieldInput[];
	
	    static createFrom(source: any = {}) {
	        return new RegenerateBodyInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.projectID = source["projectID"];
	        this.body = source["body"];
	        this.fields = this.convertValues(source["fields"], RegenerateFieldInput);
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
	
	export class SaveCapturesInput {
	    projectID: string;
	    endpointKey: string;
	    captures: EndpointCaptureDTO[];
	
	    static createFrom(source: any = {}) {
	        return new SaveCapturesInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.projectID = source["projectID"];
	        this.endpointKey = source["endpointKey"];
	        this.captures = this.convertValues(source["captures"], EndpointCaptureDTO);
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
	export class SaveCollectionInput {
	    id?: string;
	    projectID: string;
	    name: string;
	    description?: string;
	    sortOrder?: number;
	    items: CollectionItemDTO[];
	
	    static createFrom(source: any = {}) {
	        return new SaveCollectionInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.projectID = source["projectID"];
	        this.name = source["name"];
	        this.description = source["description"];
	        this.sortOrder = source["sortOrder"];
	        this.items = this.convertValues(source["items"], CollectionItemDTO);
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
	export class SaveTestsInput {
	    projectID: string;
	    endpointKey: string;
	    tests: EndpointTestDTO[];
	
	    static createFrom(source: any = {}) {
	        return new SaveTestsInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.projectID = source["projectID"];
	        this.endpointKey = source["endpointKey"];
	        this.tests = this.convertValues(source["tests"], EndpointTestDTO);
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
	export class snapshotSchemaField {
	    name: string;
	    type: string;
	    required?: boolean;
	
	    static createFrom(source: any = {}) {
	        return new snapshotSchemaField(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.type = source["type"];
	        this.required = source["required"];
	    }
	}
	export class snapshotEndpoint {
	    method: string;
	    path: string;
	    handler?: string;
	    middleware?: string[];
	    authRole?: string;
	    schemaHash?: string;
	    schemaFields?: snapshotSchemaField[];
	
	    static createFrom(source: any = {}) {
	        return new snapshotEndpoint(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.method = source["method"];
	        this.path = source["path"];
	        this.handler = source["handler"];
	        this.middleware = source["middleware"];
	        this.authRole = source["authRole"];
	        this.schemaHash = source["schemaHash"];
	        this.schemaFields = this.convertValues(source["schemaFields"], snapshotSchemaField);
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
	export class SnapshotDiffEntry {
	    method: string;
	    path: string;
	    kind: string;
	    changes?: string[];
	    authRole?: string;
	    handler?: string;
	    previous?: snapshotEndpoint;
	    current?: snapshotEndpoint;
	
	    static createFrom(source: any = {}) {
	        return new SnapshotDiffEntry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.method = source["method"];
	        this.path = source["path"];
	        this.kind = source["kind"];
	        this.changes = source["changes"];
	        this.authRole = source["authRole"];
	        this.handler = source["handler"];
	        this.previous = this.convertValues(source["previous"], snapshotEndpoint);
	        this.current = this.convertValues(source["current"], snapshotEndpoint);
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
	export class SnapshotDiff {
	    id: string;
	    // Go type: time
	    scannedAt: any;
	    previousID?: string;
	    added: SnapshotDiffEntry[];
	    removed: SnapshotDiffEntry[];
	    changed: SnapshotDiffEntry[];
	
	    static createFrom(source: any = {}) {
	        return new SnapshotDiff(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.scannedAt = this.convertValues(source["scannedAt"], null);
	        this.previousID = source["previousID"];
	        this.added = this.convertValues(source["added"], SnapshotDiffEntry);
	        this.removed = this.convertValues(source["removed"], SnapshotDiffEntry);
	        this.changed = this.convertValues(source["changed"], SnapshotDiffEntry);
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
	
	export class SnapshotSummary {
	    id: string;
	    projectID: string;
	    endpointCount: number;
	    // Go type: time
	    scannedAt: any;
	    added: number;
	    removed: number;
	    changed: number;
	
	    static createFrom(source: any = {}) {
	        return new SnapshotSummary(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.projectID = source["projectID"];
	        this.endpointCount = source["endpointCount"];
	        this.scannedAt = this.convertValues(source["scannedAt"], null);
	        this.added = source["added"];
	        this.removed = source["removed"];
	        this.changed = source["changed"];
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
	
	export class FormattedTraceFrame {
	    file?: string;
	    line?: number;
	    function?: string;
	
	    static createFrom(source: any = {}) {
	        return new FormattedTraceFrame(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.file = source["file"];
	        this.line = source["line"];
	        this.function = source["function"];
	    }
	}
	export class FormattedException {
	    message: string;
	    class?: string;
	    file?: string;
	    line?: number;
	    trace?: FormattedTraceFrame[];
	    extra?: Record<string, any>;
	
	    static createFrom(source: any = {}) {
	        return new FormattedException(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.message = source["message"];
	        this.class = source["class"];
	        this.file = source["file"];
	        this.line = source["line"];
	        this.trace = this.convertValues(source["trace"], FormattedTraceFrame);
	        this.extra = source["extra"];
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
	
	export class Timeline {
	    dnsMs: number;
	    connectMs: number;
	    tlsMs: number;
	    ttfbMs: number;
	    downloadMs: number;
	
	    static createFrom(source: any = {}) {
	        return new Timeline(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.dnsMs = source["dnsMs"];
	        this.connectMs = source["connectMs"];
	        this.tlsMs = source["tlsMs"];
	        this.ttfbMs = source["ttfbMs"];
	        this.downloadMs = source["downloadMs"];
	    }
	}
	export class Response {
	    status: number;
	    statusText: string;
	    headers?: Record<string, Array<string>>;
	    body?: string;
	    durationMs: number;
	    sizeBytes: number;
	    timeline?: Timeline;
	
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
	        this.timeline = this.convertValues(source["timeline"], Timeline);
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

