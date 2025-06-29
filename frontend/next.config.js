module.exports = {
    output: 'standalone',
    images: {
        dangerouslyAllowSVG: true,
        domains: ['localhost', '127.0.0.1', 'caddy'],
    },
    async rewrites() {
        const backendURL = process.env.BACKEND_URL || 'http://localhost:8080'
        const s3URL = process.env.S3_URL || 'http://localhost:9000'
        return [
            {
                source: '/api/:path*',
                destination: `${backendURL}/:path*`,
            },
            {
                source: '/s3/:path*',
                destination: `${s3URL}/:path*`,
            },
        ]
    },
};