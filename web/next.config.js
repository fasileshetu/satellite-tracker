/** @type {import('next').NextConfig} */
const nextConfig = {
  reactStrictMode: true,
  // The gRPC client (@grpc/grpc-js) is only ever imported from server-side
  // route handlers (app/api/telemetry/.../route.ts), but Next still tries
  // to trace/bundle it for the client graph unless it's marked external.
  serverExternalPackages: ["@grpc/grpc-js", "@grpc/proto-loader"],
};

module.exports = nextConfig;
