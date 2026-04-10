import van, { State } from 'vanjs-core'
import { Route, goto, now } from 'vanjs-router'
import { checkLogin, GLOBAL_HAS_LOGIN, GLOBAL_HIDE_PAGE, ResJSON, VanComponent } from '../mixin'
import { deleteTask, getActiveTask, getTaskList, cancelTask } from './data'
import { TaskInDB, TaskStatus } from '../work/type'
import { LoadingBox } from '../view'
import { PlayerModalComp } from './playerModal'

const { div, span, button, input } = van.tags

const { svg, path } = van.tags('http://www.w3.org/2000/svg')

export class TaskRoute implements VanComponent {
    element: HTMLElement
    /** 包含视频播放器的模态框 */
    playerModalComp = new PlayerModalComp()

    loading = van.state(false)

    /** 选中的任务ID */
    selectedIds: State<Set<number>> = van.state(new Set())

    taskList: State<(TaskInDB & {
        /** 音频下载进度百分比 */
        audioProgress: State<number>
        /** 视频下载进度百分比 */
        videoProgress: State<number>
        /** 合并进度百分比 */
        mergeProgress: State<number>
        /** 任务状态 */
        statusState: State<TaskStatus>
        /** 是否正在删除 */
        deleting: State<boolean>
    })[]> = van.state([])

    hasRunningTasks = van.derive(() =>
        this.taskList.val.some(task => task.statusState.val === 'running' || task.statusState.val === 'waiting')
    )

    hasSelectedTasks = van.derive(() => this.selectedIds.val.size > 0)

    constructor() {
        this.element = this.Root()
    }

    Root() {
        const _that = this
        return Route({
            rule: 'task',
            Loader() {
                return div(
                    () => _that.loading.val ? LoadingBox() : '',
                    () => _that.loading.val ? '' : div({ class: 'vstack gap-2' },
                        div({ class: 'hstack justify-content-end gap-2' },
                            button({
                                class: 'btn btn-outline-primary btn-sm',
                                hidden: () => !_that.hasSelectedTasks.val,
                                onclick() {
                                    const ids = Array.from(_that.selectedIds.val)
                                    if (!confirm(`确定要取消选中的 ${ids.length} 个任务吗？`)) return
                                    cancelTask(ids).then(() => {
                                        _that.selectedIds.val = new Set()
                                    }).catch(error => alert(error.message))
                                }
                            }, () => `取消选中 (${_that.selectedIds.val.size})`),
                            button({
                                class: 'btn btn-outline-danger btn-sm',
                                hidden: () => !_that.hasRunningTasks.val,
                                onclick() {
                                    if (!confirm('确定要取消所有正在进行的任务吗？')) return
                                    const runningIds = _that.taskList.val
                                        .filter(task => task.statusState.val === 'running' || task.statusState.val === 'waiting')
                                        .map(task => task.id)
                                    cancelTask(runningIds).then(() => {
                                        _that.selectedIds.val = new Set()
                                    }).catch(error => alert(error.message))
                                }
                            }, '取消所有任务')
                        ),
                        div({ class: 'list-group' },
                            _that.taskList.val.map(task => {
                                const ext = task.downloadType === 'audio' ? '.m4a' : '.mp4'
                                const filename = `${task.title} ${btoa(task.id.toString()).replace(/=/g, '')}${ext}`
                                const isSelected = van.derive(() => _that.selectedIds.val.has(task.id))
                                return div({
                                    class: () => `list-group-item p-0 hstack user-select-none ${task.statusState.val != 'done' && task.statusState.val != 'error' || task.deleting.val ? 'disabled' : ''}`,
                                    hidden: task.deleting,
                                },
                                // 复选框
                                input({
                                    type: 'checkbox',
                                    class: 'form-check-input ms-2',
                                    checked: isSelected,
                                    style: 'cursor: pointer;',
                                    onclick() {
                                        const newSet = new Set(_that.selectedIds.val)
                                        if (newSet.has(task.id)) {
                                            newSet.delete(task.id)
                                        } else {
                                            newSet.add(task.id)
                                        }
                                        _that.selectedIds.val = newSet
                                    }
                                }),
                                div({
                                    class: 'vstack gap-2 py-2 px-3 flex-fill',
                                    style: `cursor: pointer;`,
                                    onclick() {
                                        const src = `/api/downloadVideo?path=${encodeURIComponent(
                                            `${task.folder}\\${filename}`
                                        )}`
                                        if (task.statusState.val != 'done') return
                                        _that.playerModalComp.open(src, task.title, task.downloadType === 'audio' ? 'audio' : 'video')
                                    }
                                },
                                    div({
                                        class: () => `
                                        ${task.statusState.val == 'error' ? 'text-danger' : ''}
                                        ${task.statusState.val == 'waiting' || task.statusState.val == 'running'
                                                ? 'text-primary' : ''}`
                                    },
                                        div(
                                            span({
                                                class: `me-2 badge ${task.downloadType === 'audio' ? 'bg-success' : 'bg-primary'}`,
                                                title: task.downloadType === 'audio' ? '音频' : '视频'
                                            }, task.downloadType === 'audio' ? 'A' : 'V'),
                                            span({}, filename),
                                        )),
                                    div({ class: 'text-secondary small' },
                                        () => {
                                            if (task.statusState.val == 'waiting') return '等待下载'
                                            if (task.statusState.val == 'error') return '下载失败'
                                            if (task.statusState.val == 'done') return task.folder
                                            if (task.videoProgress.val == 0) {
                                                return `正在下载音频 (${(task.audioProgress.val * 100).toFixed(2)}%)`
                                            } else if (task.mergeProgress.val == 0) {
                                                return `正在下载视频 (${(task.videoProgress.val * 100).toFixed(2)}%)`
                                            } else if (task.statusState.val == 'running') {
                                                return `正在合并音视频 (${(task.mergeProgress.val * 100).toFixed(2)}%)`
                                            } else {
                                                return task.folder
                                            }
                                        }
                                    ),
                                    div({
                                        class: `progress`,
                                        style: `height: 5px`,
                                        hidden: () => task.statusState.val == 'done' || task.statusState.val == 'error'
                                    },
                                        div({
                                            class: () => `progress-bar progress-bar-striped progress-bar-animated bg-${(() => {
                                                if (task.videoProgress.val == 0) return 'primary'
                                                if (task.mergeProgress.val == 0) return 'success'
                                                else return 'info'
                                            })()}`,
                                            style: () => {
                                                let width = 0
                                                if (task.videoProgress.val == 0) width = task.audioProgress.val * 100
                                                else if (task.mergeProgress.val == 0) width = task.videoProgress.val * 100
                                                else width = task.mergeProgress.val * 100
                                                return `width: ${width}%`
                                            }
                                        }),
                                    )
                                ),
                                div({ class: 'me-2 hstack gap-1' },
                                    div({
                                        class: 'hover-btn',
                                        title: '取消任务',
                                        hidden: () => task.statusState.val !== 'running' && task.statusState.val !== 'waiting',
                                        onclick() {
                                            if (!confirm('确定要取消这个任务吗？')) return
                                            cancelTask([task.id]).catch(error => alert(error.message))
                                        }
                                    }, _that.CancelSVG()),
                                    div({
                                        class: 'hover-btn',
                                        title: '删除视频',
                                        hidden: () => task.statusState.val != 'done' && task.statusState.val != 'error',
                                        onclick() {
                                            task.deleting.val = true
                                            deleteTask(task.id).then(() => {
                                                _that.taskList.val = _that.taskList.val.filter(taskInDB => taskInDB.id != task.id)
                                            }).catch(error => {
                                                alert(error.message)
                                            })
                                        }
                                    }, _that.DeleteSVG())
                                ))
                            })
                        )
                    )
                )
            },
            async onFirst() {
                if (!await checkLogin()) return
            },
            async onLoad() {
                if (!GLOBAL_HAS_LOGIN.val) return goto('login')
                _that.loading.val = true

                getTaskList(0, 360).then(taskList => {
                    if (!taskList) return
                    _that.taskList.val = taskList.map(task => ({
                        ...task,
                        audioProgress: van.state(1),
                        videoProgress: van.state(1),
                        mergeProgress: van.state(1),
                        statusState: van.state(task.status),
                        deleting: van.state(false)
                    }))

                    const refresh = async () => {
                        const activeTaskList = await getActiveTask()
                        if (!activeTaskList) return false
                        setTimeout(() => {
                            _that.loading.val = false
                        }, 200)

                        _that.taskList.val.forEach(taskInDB => {
                            activeTaskList.forEach(task => {
                                if (taskInDB.id == task.id) {
                                    taskInDB.audioProgress.val = task.audioProgress
                                    taskInDB.videoProgress.val = task.videoProgress
                                    taskInDB.mergeProgress.val = task.mergeProgress
                                    taskInDB.statusState.val = task.status
                                }
                            })
                        })
                        if (activeTaskList.filter(task => task.status == 'running' || task.status == 'waiting').length == 0) {
                            clearInterval(timer)
                            clearInterval(helper)
                        }
                        return true
                    }

                    refresh()

                    let timer = setInterval(() => {
                        refresh()
                    }, 1000)
                    let helper = setInterval(() => {
                        if (now.val.split('/')[0] != 'task') {
                            clearInterval(helper)
                            clearInterval(timer)
                        }
                    })
                })
            },
        })
    }

    DeleteSVG() {
        return svg({ style: `width: 1em; height: 1em`, fill: "currentColor", class: "bi bi-trash3", viewBox: "0 0 16 16" },
            path({ "d": "M6.5 1h3a.5.5 0 0 1 .5.5v1H6v-1a.5.5 0 0 1 .5-.5M11 2.5v-1A1.5 1.5 0 0 0 9.5 0h-3A1.5 1.5 0 0 0 5 1.5v1H1.5a.5.5 0 0 0 0 1h.538l.853 10.66A2 2 0 0 0 4.885 16h6.23a2 2 0 0 0 1.994-1.84l.853-10.66h.538a.5.5 0 0 0 0-1zm1.958 1-.846 10.58a1 1 0 0 1-.997.92h-6.23a1 1 0 0 1-.997-.92L3.042 3.5zm-7.487 1a.5.5 0 0 1 .528.47l.5 8.5a.5.5 0 0 1-.998.06L5 5.03a.5.5 0 0 1 .47-.53Zm5.058 0a.5.5 0 0 1 .47.53l-.5 8.5a.5.5 0 1 1-.998-.06l.5-8.5a.5.5 0 0 1 .528-.47M8 4.5a.5.5 0 0 1 .5.5v8.5a.5.5 0 0 1-1 0V5a.5.5 0 0 1 .5-.5" }),
        )
    }

    CancelSVG() {
        return svg({ style: `width: 1em; height: 1em`, fill: "currentColor", class: "bi bi-x-circle", viewBox: "0 0 16 16" },
            path({ "d": "M8 15A7 7 0 1 1 8 1a7 7 0 0 1 0 14m0 1A8 8 0 1 0 8 0a8 8 0 0 0 0 16" }),
            path({ "d": "M4.646 4.646a.5.5 0 0 1 .708 0L8 7.293l2.646-2.647a.5.5 0 0 1 .708.708L8.707 8l2.647 2.646a.5.5 0 0 1-.708.708L8 8.707l-2.646 2.647a.5.5 0 0 1-.708-.708L7.293 8 4.646 5.354a.5.5 0 0 1 0-.708" }),
        )
    }
}

export default () => new TaskRoute().element