export namespace app {
	
	export class Bucket {
	    Name: string;
	    CreationDate: string;
	    Location: string;
	
	    static createFrom(source: any = {}) {
	        return new Bucket(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Name = source["Name"];
	        this.CreationDate = source["CreationDate"];
	        this.Location = source["Location"];
	    }
	}
	export class OBSObject {
	    Key: string;
	    Size: number;
	    LastModified: string;
	    StorageClass: string;
	
	    static createFrom(source: any = {}) {
	        return new OBSObject(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Key = source["Key"];
	        this.Size = source["Size"];
	        this.LastModified = source["LastModified"];
	        this.StorageClass = source["StorageClass"];
	    }
	}
	export class ListObjectsResponse {
	    objects: OBSObject[];
	    folders: string[];
	
	    static createFrom(source: any = {}) {
	        return new ListObjectsResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.objects = this.convertValues(source["objects"], OBSObject);
	        this.folders = source["folders"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice) {
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
	
	export class PreviewFileResponse {
	    content: string;
	    contentType: string;
	    fileName: string;
	    size: number;
	
	    static createFrom(source: any = {}) {
	        return new PreviewFileResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.content = source["content"];
	        this.contentType = source["contentType"];
	        this.fileName = source["fileName"];
	        this.size = source["size"];
	    }
	}

}

export namespace connection {
	
	export class Connection {
	    id: string;
	    name: string;
	    provider: string;
	    accessKeyId: string;
	    secretAccessKey: string;
	    region: string;
	    endpoint: string;
	    status: string;
	    delimiter: string;
	    extraConfig: string;
	    // Go type: time
	    createdAt: any;
	    // Go type: time
	    updatedAt: any;
	
	    static createFrom(source: any = {}) {
	        return new Connection(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.provider = source["provider"];
	        this.accessKeyId = source["accessKeyId"];
	        this.secretAccessKey = source["secretAccessKey"];
	        this.region = source["region"];
	        this.endpoint = source["endpoint"];
	        this.status = source["status"];
	        this.delimiter = source["delimiter"];
	        this.extraConfig = source["extraConfig"];
	        this.createdAt = this.convertValues(source["createdAt"], null);
	        this.updatedAt = this.convertValues(source["updatedAt"], null);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice) {
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

export namespace theme {
	
	export class BackgroundInfo {
	    imageBase64: string;
	    opacity: number;
	    hasImage: boolean;
	
	    static createFrom(source: any = {}) {
	        return new BackgroundInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.imageBase64 = source["imageBase64"];
	        this.opacity = source["opacity"];
	        this.hasImage = source["hasImage"];
	    }
	}
	export class ThemeConfig {
	    id: string;
	    name: string;
	    description: string;
	    preview: string;
	    glassEnabled: boolean;
	    variables: {[key: string]: any};
	
	    static createFrom(source: any = {}) {
	        return new ThemeConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.description = source["description"];
	        this.preview = source["preview"];
	        this.glassEnabled = source["glassEnabled"];
	        this.variables = source["variables"];
	    }
	}

}

