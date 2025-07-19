export namespace main {
	
	export class LoginRequest {
	    username: string;
	    password: string;
	
	    static createFrom(source: any = {}) {
	        return new LoginRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.username = source["username"];
	        this.password = source["password"];
	    }
	}
	export class LoginResponse {
	    success: boolean;
	    message: string;
	    requiresTwoFactor: boolean;
	    userDisplayName?: string;
	    errors?: string[];
	
	    static createFrom(source: any = {}) {
	        return new LoginResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.message = source["message"];
	        this.requiresTwoFactor = source["requiresTwoFactor"];
	        this.userDisplayName = source["userDisplayName"];
	        this.errors = source["errors"];
	    }
	}
	export class TwoFactorRequest {
	    code: string;
	    isRecoveryCode: boolean;
	
	    static createFrom(source: any = {}) {
	        return new TwoFactorRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.code = source["code"];
	        this.isRecoveryCode = source["isRecoveryCode"];
	    }
	}
	export class UploadRequest {
	    imagePath: string;
	    note: string;
	    worldId: string;
	    worldName: string;
	    noResize: boolean;
	
	    static createFrom(source: any = {}) {
	        return new UploadRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.imagePath = source["imagePath"];
	        this.note = source["note"];
	        this.worldId = source["worldId"];
	        this.worldName = source["worldName"];
	        this.noResize = source["noResize"];
	    }
	}
	export class UploadResponse {
	    success: boolean;
	    message: string;
	    fileId?: string;
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new UploadResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.message = source["message"];
	        this.fileId = source["fileId"];
	        this.error = source["error"];
	    }
	}

}

