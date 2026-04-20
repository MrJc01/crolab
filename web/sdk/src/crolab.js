/**
 * Crolab SDK - Core Javascript Framework
 * Agnostic API wrapper for the Crolab Ecosystem.
 */

export class CrolabSDK {
  constructor(apiUrl = '') {
    if (!apiUrl && typeof window !== 'undefined') {
      apiUrl = window.location.origin;
    }
    this.API = apiUrl;
    
    // In browser environments we use localStorage, in Node we use mock memory.
    this.isBrowser = typeof window !== 'undefined' && typeof window.localStorage !== 'undefined';
    this._memToken = '';
  }

  get token() {
    if (this.isBrowser) {
      return window.localStorage.getItem('admin_token') || '';
    }
    return this._memToken;
  }

  set token(val) {
    if (this.isBrowser) {
      window.localStorage.setItem('admin_token', val);
    }
    this._memToken = val;
  }

  async _api(method, path, body) {
    const opts = { method, headers: { 'Content-Type': 'application/json' } };
    const t = this.token;
    if (t) {
      opts.headers['Authorization'] = t;
    }
    if (body) {
      opts.body = JSON.stringify(body);
    }
    try {
      const res = await fetch(this.API + path, opts);
      const data = await res.json().catch(() => ({}));
      return { status: res.status, data };
    } catch (e) {
      return { status: 0, data: { error: e.message } };
    }
  }

  async login(email, password) {
    const { status, data } = await this._api('POST', '/auth/login', { email, password });
    if (status === 200 && data.token) {
      this.token = data.token;
    }
    return { status, data };
  }

  async logout() {
    this.token = '';
    return true;
  }

  async getDashboard() {
    return await this._api('GET', '/admin/dashboard');
  }

  async getPlans() {
    return await this._api('GET', '/admin/plans');
  }

  async getMachines() {
    return await this._api('GET', '/admin/machines');
  }

  async syncCloud() {
    return await this._api('POST', '/admin/providers/sync');
  }
}

// Instantiate default global SDK
export const crolab = new CrolabSDK();
