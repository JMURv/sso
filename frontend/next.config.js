module.exports = {
    output: 'standalone',
    images: {
        domains: ['localhost', '127.0.0.1'],
    },
    async rewrites() {
        return [
            {
                source: '/api/:path*',
                destination: `${process.env.BACKEND_URL}/api/:path*`,
            },
        ]
    },
};