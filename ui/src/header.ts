import van from 'vanjs-core'
import { now } from 'vanjs-router'
import { GLOBAL_HAS_LOGIN } from './mixin'
import { checkFFmpeg } from './setting/data'

const { a, button, div, span } = van.tags

export const GLOBAL_FFMPEG_STATUS = {
    available: van.state(false),
    version: van.state('')
}

// 初始化检查 ffmpeg
checkFFmpeg().then(status => {
    GLOBAL_FFMPEG_STATUS.available.val = status.available
    GLOBAL_FFMPEG_STATUS.version.val = status.version
})

export default () => {
    const classStr = (name: string) => van.derive(() => `text-nowrap nav-link ${now.val.split('/')[0] == name ? 'active' : ''}`)

    return div({ class: 'hstack gap-4' },
        div({ class: 'fs-4 fw-bold text-nowrap' }, 'Bilidown'),
        div({ class: 'nav nav-underline flex-nowrap overflow-auto' },
            div({ class: 'nav-item', hidden: () => !GLOBAL_HAS_LOGIN.val },
                a({ class: classStr('work'), href: '#/work' }, '视频解析')
            ),
            div({ class: 'nav-item', hidden: () => !GLOBAL_HAS_LOGIN.val },
                a({ class: classStr('task'), href: '#/task' }, '任务列表')
            ),
            div({ class: 'nav-item', hidden: () => !GLOBAL_HAS_LOGIN.val },
                a({ class: classStr('setting'), href: '#/setting' }, '设置中心')
            ),
            div({ class: 'nav-item', hidden: GLOBAL_HAS_LOGIN },
                a({ class: classStr('login'), href: '#/login' }, '扫码登录')
            ),
        ),
        div({ class: 'ms-auto hstack gap-2' },
            () => GLOBAL_FFMPEG_STATUS.available.val
                ? span({ class: 'badge bg-success' }, `FFmpeg (${GLOBAL_FFMPEG_STATUS.version.val})`)
                : a({
                    class: 'badge bg-danger text-decoration-none',
                    href: 'https://www.ffmpeg.org/',
                    target: '_blank'
                }, 'FFmpeg 未安装')
        )
    )
}