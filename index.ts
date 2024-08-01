import * as pulumi from "@pulumi/pulumi";

console.log(process.env)
console.log(process.env.secret?.split(""))

// Export the DNS name of the bucket
export const bucketName = "dummy";
