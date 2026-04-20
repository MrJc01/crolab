/**
 * Crolab SDK (JavaScript / Node.js Compatible)
 * Universal bridge to execute and connect on remote SRE GPU clusters of Crolab.
 */

class CrolabSDK {
    /**
     * Instancia o SDK Crolab (Web ou Node).
     * @param {string} cloudUrl - URL base do laboratório/nuvem. Padrão "http://localhost:8844"
     * @param {string} token - Bearer Token de conta Crolab.
     */
    constructor(cloudUrl = "http://localhost:8844", token = "") {
        this.cloudUrl = cloudUrl;
        this.token = token;
    }

    /**
     * Roda um bloco de código cru interativamente na nuvem ou Node Pool SRE, abstraindo WebSocket e Docker.
     * @param {string} code - O código em texto.
     * @param {Object} options - language ('python', 'js', 'bash'), planId
     * @returns {Promise<Object>} Resposta JSON resolvida de output padrão e stderror.
     */
    async run(code, options = {}) {
        const language = options.language || 'python';
        const planId = options.planId || '';
        const body = {
            language: language,
            code: code,
            plan_id: planId
        };

        const response = await fetch(`${this.cloudUrl}/client/run/inline`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': this.token
            },
            body: JSON.stringify(body)
        });

        if (!response.ok) {
            const errData = await response.json();
            throw new Error(errData.error || response.statusText);
        }

        return await response.json(); 
    }

    /**
     * Submete um diretório para deploy em massa baseando no Crolab Local Engine (requer childProc backend proxy num fs local).
     * Essa função é opcionalmente suportada via node-Crolab adapter se houver file streams.
     */
    async submitBundle(blobArchive) {
         throw new Error("SDK uploadStream requires File API capabilities - Unimplemented");
    }
}

if (typeof module !== 'undefined' && module.exports) {
    module.exports = { CrolabSDK };
} else {
    window.CrolabSDK = CrolabSDK;
}
