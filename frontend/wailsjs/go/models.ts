export namespace models {
	
	export class TemplateOption {
	    label: string;
	    value: string;
	
	    static createFrom(source: any = {}) {
	        return new TemplateOption(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.label = source["label"];
	        this.value = source["value"];
	    }
	}
	export class TemplateParam {
	    name: string;
	    type: string;
	    description: string;
	    options: TemplateOption[];
	
	    static createFrom(source: any = {}) {
	        return new TemplateParam(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.type = source["type"];
	        this.description = source["description"];
	        this.options = this.convertValues(source["options"], TemplateOption);
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
	export class Command {
	    id: number;
	    name: string;
	    content: string;
	    description: string;
	    groupId?: number;
	    workspaceId: number;
	    templateParams: TemplateParam[];
	
	    static createFrom(source: any = {}) {
	        return new Command(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.content = source["content"];
	        this.description = source["description"];
	        this.groupId = source["groupId"];
	        this.workspaceId = source["workspaceId"];
	        this.templateParams = this.convertValues(source["templateParams"], TemplateParam);
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
	export class CommandGroup {
	    id: number;
	    name: string;
	    workspaceId: number;
	
	    static createFrom(source: any = {}) {
	        return new CommandGroup(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.workspaceId = source["workspaceId"];
	    }
	}
	export class RecentPath {
	    id: number;
	    workspaceId: number;
	    path: string;
	    position: number;
	
	    static createFrom(source: any = {}) {
	        return new RecentPath(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.workspaceId = source["workspaceId"];
	        this.path = source["path"];
	        this.position = source["position"];
	    }
	}
	export class IgnoreRule {
	    pattern: string;
	    isRegex: boolean;
	
	    static createFrom(source: any = {}) {
	        return new IgnoreRule(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.pattern = source["pattern"];
	        this.isRegex = source["isRegex"];
	    }
	}
	export class Workspace {
	    id: number;
	    name: string;
	    path: string;
	    ignoredCommands: IgnoreRule[];
	
	    static createFrom(source: any = {}) {
	        return new Workspace(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.path = source["path"];
	        this.ignoredCommands = this.convertValues(source["ignoredCommands"], IgnoreRule);
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
	export class DatabaseBackup {
	    version: string;
	    workspaces: Workspace[];
	    groups: CommandGroup[];
	    commands: Command[];
	    recentPaths: RecentPath[];
	
	    static createFrom(source: any = {}) {
	        return new DatabaseBackup(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.version = source["version"];
	        this.workspaces = this.convertValues(source["workspaces"], Workspace);
	        this.groups = this.convertValues(source["groups"], CommandGroup);
	        this.commands = this.convertValues(source["commands"], Command);
	        this.recentPaths = this.convertValues(source["recentPaths"], RecentPath);
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
	
	
	
	
	
	export class WorkspaceExport {
	    version: string;
	    name: string;
	    groups: CommandGroup[];
	    commands: Command[];
	    ignored: IgnoreRule[];
	
	    static createFrom(source: any = {}) {
	        return new WorkspaceExport(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.version = source["version"];
	        this.name = source["name"];
	        this.groups = this.convertValues(source["groups"], CommandGroup);
	        this.commands = this.convertValues(source["commands"], Command);
	        this.ignored = this.convertValues(source["ignored"], IgnoreRule);
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

