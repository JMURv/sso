module.exports = {
    output: 'standalone',
    images: {
        dangerouslyAllowSVG: true,
        domains: ['localhost', '127.0.0.1', 'caddy'],
    },
    async rewrites() {
        return [
            {
                source: '/api/:path*',
                destination: `${process.env.BACKEND_URL ? process.env.BACKEND_URL : 'http://localhost:8080'}/:path*`,
            },
            {
                source: '/s3/:path*',
                destination: `${process.env.S3_URL ? process.env.S3_URL : 'http://localhost:9000'}/:path*`,
            },
        ]
    },
};