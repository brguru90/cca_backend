module.exports = {
    apps: [
        {
            name: "api_server",
            "script": "./go_server",
            "exec_interpreter": "none",
            watch: false,
            exec_mode: "fork_mode",
            instances: 1,
            env_pm2: {
                "NODE_ENV": "production"
            }
        },

    ]
}