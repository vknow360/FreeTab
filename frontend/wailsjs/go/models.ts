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

