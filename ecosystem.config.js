module.exports = {
    apps: [
        {
            name: "api_server",
            "script": process.env?.IS_VIDEO_PROCESSING!="true"?"./go_server":"./go_server -micro_service video_processing",
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