export namespace device {
	
	export class DeviceInfo {
	    connected: boolean;
	    state: string;
	    serial: string;
	    model: string;
	    androidRelease: string;
	    vendorSdk: string;
	    rootGranted: boolean;
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new DeviceInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.connected = source["connected"];
	        this.state = source["state"];
	        this.serial = source["serial"];
	        this.model = source["model"];
	        this.androidRelease = source["androidRelease"];
	        this.vendorSdk = source["vendorSdk"];
	        this.rootGranted = source["rootGranted"];
	        this.error = source["error"];
	    }
	}
	export class DriverItem {
	    name: string;
	    installed: boolean;
	    deviceSeen: boolean;
	    service: string;
	    status: string;
	    details: string;
	
	    static createFrom(source: any = {}) {
	        return new DriverItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.installed = source["installed"];
	        this.deviceSeen = source["deviceSeen"];
	        this.service = source["service"];
	        this.status = source["status"];
	        this.details = source["details"];
	    }
	}
	export class DriverReport {
	    ready: boolean;
	    adbDriver: DriverItem;
	    fastbootDriver: DriverItem;
	    mtkDriver: DriverItem;
	    problemFound: boolean;
	    summary: string;
	    downloadUrl: string;
	    actionNeeded: string;
	
	    static createFrom(source: any = {}) {
	        return new DriverReport(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ready = source["ready"];
	        this.adbDriver = this.convertValues(source["adbDriver"], DriverItem);
	        this.fastbootDriver = this.convertValues(source["fastbootDriver"], DriverItem);
	        this.mtkDriver = this.convertValues(source["mtkDriver"], DriverItem);
	        this.problemFound = source["problemFound"];
	        this.summary = source["summary"];
	        this.downloadUrl = source["downloadUrl"];
	        this.actionNeeded = source["actionNeeded"];
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
	export class ToolPaths {
	    adb: string;
	    fastboot: string;
	
	    static createFrom(source: any = {}) {
	        return new ToolPaths(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.adb = source["adb"];
	        this.fastboot = source["fastboot"];
	    }
	}

}

export namespace main {
	
	export class DeviceState {
	    mode: string;
	    adbInfo?: device.DeviceInfo;
	    fastbootSerial?: string;
	
	    static createFrom(source: any = {}) {
	        return new DeviceState(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.mode = source["mode"];
	        this.adbInfo = this.convertValues(source["adbInfo"], device.DeviceInfo);
	        this.fastbootSerial = source["fastbootSerial"];
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

