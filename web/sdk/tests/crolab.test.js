import { describe, it, expect, vi, beforeEach } from 'vitest';
import { CrolabSDK } from '../src/crolab.js';

describe('Crolab SDK', () => {
  let sdk;

  beforeEach(() => {
    sdk = new CrolabSDK('http://mockapi');
    global.fetch = vi.fn();
  });

  it('should initialize with empty token', () => {
    expect(sdk.token).toBe('');
  });

  it('should handle successful login and store token', async () => {
    global.fetch.mockResolvedValueOnce({
      status: 200,
      json: async () => ({ token: 'mock_bearer_token', user: 'admin' })
    });

    const res = await sdk.login('admin@crolab.local', 'password123');
    
    expect(global.fetch).toHaveBeenCalledWith('http://mockapi/auth/login', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ email: 'admin@crolab.local', password: 'password123' })
    });
    
    expect(res.status).toBe(200);
    expect(res.data.token).toBe('mock_bearer_token');
    expect(sdk.token).toBe('mock_bearer_token');
  });

  it('should handle network crash gracefully safely returning status 0', async () => {
    global.fetch.mockRejectedValueOnce(new Error('Network Offline'));

    const res = await sdk.getMachines();
    
    expect(res.status).toBe(0);
    expect(res.data.error).toBe('Network Offline');
  });

  it('should automatically inject Bearer Authorization header if token exists', async () => {
    sdk.token = 'cr0_secretToken';
    global.fetch.mockResolvedValueOnce({
      status: 200,
      json: async () => ({ status: 'success' })
    });

    await sdk.syncCloud();

    expect(global.fetch).toHaveBeenCalledWith('http://mockapi/admin/providers/sync', {
      method: 'POST',
      headers: { 
        'Content-Type': 'application/json',
        'Authorization': 'cr0_secretToken'
      }
    });
  });
});
