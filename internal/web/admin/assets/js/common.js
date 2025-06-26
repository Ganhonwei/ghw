var MSG_OK = 0;
var MSG_ERR = -1;
var MSG_REDIRECT = -2;
	
jQuery(function($){
	$(".pagination-jump a").click(function (e) {
        var pages = document.getElementById("input_page")
        console.log("输入页面："+pages.innerText)
    })
	//ajax提交
	$('.ajax-form').on('submit', function() {
		var url = $('.ajax-form').attr('action') + "?_t=" + Math.random();
		$('.alert').addClass('hide');
		$('button[type="submit"]').attr('disabled', true);
		if ($('#editor')) {
			$("input[name='editor_content']").val($('#editor').html());
		}
		$.post(url, $(".ajax-form").serialize(), function (out) {
			if (out.status == MSG_OK) { // 成功
				if (out.redirect != "") {
					window.location.href = out.redirect;
				} else {
					window.location.reload();
				}
			} else if (out.status == MSG_REDIRECT) {
				window.location.href = out.redirect;
			} else if (out.status == MSG_ERR) {
				if ($('.alert')) {
					$('.alert').removeClass('hide');
					$('.alert').html(out.msg);
				} else {
					alert(out.msg);
				}
				$('button[type="submit"]').removeAttr('disabled');
			}
		});
		return false;
	});
	
    $.datepicker.regional['zh-CN'] = {
        closeText: '关闭',
        prevText: '<上月',
        nextText: '下月>',
        currentText: '今天',
        monthNames: ['一月','二月','三月','四月','五月','六月',
            '七月','八月','九月','十月','十一月','十二月'],
        monthNamesShort: ['一','二','三','四','五','六',
            '七','八','九','十','十一','十二'],
        dayNames: ['星期日','星期一','星期二','星期三','星期四','星期五','星期六'],
        dayNamesShort: ['周日','周一','周二','周三','周四','周五','周六'],
        dayNamesMin: ['日','一','二','三','四','五','六'],
        weekHeader: '周',
        dateFormat: 'yy-mm-dd',
        firstDay: 1,
        isRTL: false,
        showMonthAfterYear: true,
        yearSuffix: '年'};
    $.datepicker.setDefaults($.datepicker.regional['zh-CN']);

    /** 日历控件 **/
    $( "#start_date, #end_date" ).datepicker({
        showOtherMonths: true,
        selectOtherMonths: false,
        dateFormat: 'yy-mm-dd',
        changeYear:true,
        changeMonth:true,
        maxDate: new Date()
    });
    $( "#start_date1, #end_date1" ).datepicker({
        showOtherMonths: true,
        selectOtherMonths: false,
        dateFormat: 'yy-mm-dd',
        changeYear:true,
        changeMonth:true,
        maxDate: new Date()
    });
    $( "#start_date2, #end_date2" ).datepicker({
        showOtherMonths: true,
        selectOtherMonths: false,
        dateFormat: 'yy-mm-dd',
        changeYear:true,
        changeMonth:true
    });

    $(function () {
        $('[data-toggle="tooltip"]').tooltip()
    })

    /*$('.delete').click(function () {
        return confirm('确定要删除这条记录吗？');
    });*/
    $.widget("ui.dialog", $.extend({}, $.ui.dialog.prototype, {
        _title: function(title) {
            var $title = this.options.title || '&nbsp;'
            if( ("title_html" in this.options) && this.options.title_html == true )
                title.html($title);
            else title.text($title);
        }
    }));
    $( ".delete_confirm" ).on('click', function(e) {
        var del_url = $(this).attr('href');
        e.preventDefault();
        $( "#dialog-confirm" ).removeClass('hide').dialog({
            resizable: false,
            width: '320',
            modal: true,
            title: "<div class='widget-header'><h4 class='smaller'><i class='ace-icon fa fa-exclamation-triangle red'></i> 删除确认</h4></div>",
            title_html: true,
            buttons: [
                {
                    html: "<i class='ace-icon fa fa-trash-o bigger-110'></i>&nbsp; 确认",
                    "class" : "btn btn-danger btn-sm",
                    click: function() {
                        $( this).dialog('close');
                        window.location.href = del_url;
                    }
                }
                ,
                {
                    html: "<i class='ace-icon fa fa-times bigger-110'></i>&nbsp; 取消",
                    "class" : "btn btn-sm",
                    click: function() {
                        $( this ).dialog( "close" );
                    }
                }
            ]
        });
    });

    const calcTableWrapHeighterDebounce = debounce(tableWrapHeighter, 500);
    window.addEventListener('resize', calcTableWrapHeighterDebounce);
    tableWrapHeighter();

    
    // 表头右键菜单处理
    // 点击其他地方时隐藏菜单
    $(document).on('click', function() {
        $('#contextMenu').hide();
    });

    // 处理菜单项点击
    $('.context-menu-item').on('click', function() {
        const action = $(this).data('action');
        if (!window.currentTh) { return }
        // console.log(window.currentTh);
        
        // 取消固定
        var key = window.location.pathname;
        if (window.stickyKeySuffix) {
            key += (":" + window.stickyKeySuffix);
        }
        var end = false;
        if (action === 'fixLeft') {
            if (window.currentTh.hasAttribute('stickyL')) {
                window.currentTh.removeAttribute('stickyL');
                window.localStorage.setItem('stickyL:' + key, -1);
                end = true;
            }
        } else if (action === 'fixRight') {
            if (window.currentTh.hasAttribute('stickyR')) {
                window.currentTh.removeAttribute('stickyR');
                window.localStorage.setItem('stickyR:' + key, -1);
                end = true;
            }
        }
        if (end) {
            tableWrapHeighter();
            window.currentTh = null;
            $('#contextMenu').hide();
            return;
        }
        // 表头
        var headers = window.currentTh.parentNode.children;

        // 判断位置是否可固定
        var stickyL = null;
        var stickyR = null;
        for (var i = 0; i < headers.length; i++) {
            if (stickyL === null && headers[i].hasAttribute('stickyL')) {
                stickyL = i;
            }
            if (stickyR === null && headers[i].hasAttribute('stickyR')) {
                stickyR = i;
            }
            if (headers[i] == window.currentTh) {
                if (action === 'fixLeft') {
                    stickyL = i;
                } else if (action === 'fixRight') {
                    stickyR = i;
                }
            }
        }
        if (stickyL !== null && stickyR !== null && stickyL >= stickyR) {
            console.log('illegal stickyLR:', stickyL, stickyR);
            window.currentTh = null;
            $('#contextMenu').hide();
            return
        }

        // 变更固定位置
        for (var i = 0; i < headers.length; i++) {
            if (action === 'fixLeft') {
                headers[i].removeAttribute('stickyL');
            } else if (action === 'fixRight') {
                headers[i].removeAttribute('stickyR');
            }
        }
        if (action === 'fixLeft') {
            window.currentTh.setAttribute('stickyL', '');
        } else if (action === 'fixRight') {
            window.currentTh.setAttribute('stickyR', '');
        }

        tableWrapHeighter();
        window.currentTh = null;
        $('#contextMenu').hide();

        // 保存固定位置记录
        var key = window.location.pathname;
        if (window.stickyKeySuffix) {
            key += (":" + window.stickyKeySuffix);
        }
        if (stickyL < 7) {
            window.localStorage.setItem('stickyL:' + key, stickyL);
        }
        if (stickyR > headers.length - 7) {
            window.localStorage.setItem('stickyR:' + key, stickyR);
        }
    });
});

function debounce(fn, delay) {
    let timer;
    return function() {
        if (timer) {
            clearTimeout(timer);
        }
        timer = setTimeout(() => {
            fn();
        }, delay);
    }
};

function tableWrapHeighter() {
    var wraponlys = document.querySelectorAll("[table-wrap-only]");
    if (wraponlys.length > 0) {
        for (var i = 0; i < wraponlys.length; i++) {
            var wrap = wraponlys[i];
            var restHeight = parseInt(wrap.getAttribute("table-wrap-only") || 0) || 0;
            var height = window.innerHeight - restHeight;
            wrap.style['max-height'] = height + "px";
        }
        return
    }

    var wraps = document.querySelectorAll("[table-wrap]");
    for (var i = 0; i < wraps.length; i++) {
        var wrap = wraps[i];
        var restHeight = parseInt(wrap.getAttribute("table-wrap") || 0) || 0;
        var height = window.innerHeight - restHeight;
        // var style = wrap.getAttribute("style") || '';
        // style += `;height:${height}px;`;
        // wrap.setAttribute("style", style);
        wrap.style['max-height'] = height + "px";
        // console.log('table wrap', wrap, height)

        // 固定头n列和尾n列
        var tables = wrap.getElementsByTagName('table');
        if (tables.length == 0) {
            continue;
        }
        for (var j = 0; j < tables.length; j++) {
            var table = tables[j];
            var rows = table.getElementsByTagName('tr');
            var columnsWidth = null;
            var headerHeight = null;
            if (!rows[0]) {
                continue
            }
            // header
            var stickyL = null;
            var stickyR = null;
            for (var l = 0; l < rows[0].children.length; l++) {
                if (l == 0 && !headerHeight) {
                    headerHeight = rows[0].children[0].scrollHeight;
                }
                if (!columnsWidth) {
                    columnsWidth = new Array(rows[0].children.length);
                }
                if (!columnsWidth[l]) {
                    columnsWidth[l] = rows[0].children[l].scrollWidth;
                }

                var col = rows[0].children[l];
                if (col.hasAttribute('stickyL')) {
                    stickyL = l;
                }
                if (col.hasAttribute('stickyR')) {
                    stickyR = l;
                }
                rows[0].children[l].oncontextmenu = function(e) {
                    e.preventDefault();
                    // showMenu(e);
                    // console.log("111", e.target)
                    window.currentTh = e.target;
                    var texts = ['向左固定', '向右固定'];
                    if (e.target.hasAttribute('stickyL')) {
                        texts[0] = '取消向左固定';
                    }
                    if (e.target.hasAttribute('stickyR')) {
                        texts[1] = '取消向右固定';
                    }
                    for (var i = 0; i < texts.length; i++) {
                        $('#contextMenu .context-menu-item').eq(i).text(texts[i]);
                    }
                    // 显示自定义菜单
                    $('#contextMenu').css({
                        display: 'block',
                        left: e.pageX,
                        top: e.pageY
                    });
                }
            }

            for (var k = 0; k < rows.length; k++) {
                var row = rows[k];
                var rowBgColor = window.getComputedStyle(row).backgroundColor;
                if (rowBgColor === 'rgba(0, 0, 0, 0)') {
                    rowBgColor = 'rgba(255, 255, 255)';
                }

                if (stickyL !== null) {
                    for (var l = 0; l <= stickyL && l < row.children.length; l++) {
                        row.children[l].style.position = 'sticky';
                        row.children[l].style['z-index'] = 1;
                        row.children[l].style['background-color'] = rowBgColor;
                        // calc left px
                        var left = 0;
                        if (l > 0) {
                            for (var m = 0; m < l && m < columnsWidth.length; m++) {
                                left += columnsWidth[m];
                            }
                        }
                        row.children[l].style.left = left + 'px';
                        // 偏移表头高度
                        if (k != 0) {
                            row.children[l].style.top = headerHeight + 'px';
                        }
                    }
                }
                
                if (stickyR !== null) {
                    for (var l = row.children.length-1; l >= stickyR && l >= 0; l--) {
                        row.children[l].style.position = 'sticky';
                        row.children[l].style['z-index'] = 1;
                        row.children[l].style['background-color'] = rowBgColor;
                        // calc right px
                        var right = 0;
                        if (l < row.children.length-1) {
                            for (var m = row.children.length-1; m > l && m >= 0; m--) {
                                right += columnsWidth[m];
                            }
                        }
                        row.children[l].style.right = right + 'px';
                        // 偏移表头高度
                        if (k != 0) {
                            row.children[l].style.top = headerHeight + 'px';
                        }
                    }
                }

                // 清除之前固定的css
                stickyL = stickyL === null ? -1 : stickyL;
                stickyR = stickyR === null ? row.children.length : stickyR;
                for (var l = stickyL+1; l < stickyR && l < row.children.length; l++) {
                    if (row.children[l].style.position === 'sticky') {
                        row.children[l].style.position = '';
                        row.children[l].style['z-index'] = '';
                        row.children[l].style.left = '';
                        row.children[l].style.right = '';
                    }
                }
            }
        }
    }
}

function _StickyMemory(keySuffix) {
    if (keySuffix === undefined || keySuffix === null) {
        keySuffix = '';
    }
    window.stickyKeySuffix = keySuffix;
    var key = window.location.pathname;
    if (window.stickyKeySuffix) {
        key += (":" + window.stickyKeySuffix);
    }
    
    var stickyL = parseInt(window.localStorage.getItem('stickyL:' + key));
    var stickyR = parseInt(window.localStorage.getItem('stickyR:' + key));
    if (!Number.isNaN(stickyL)) {
        $("[table-wrap] th[stickyL]").removeAttr('stickyL');
        if (stickyL != -1) {
            $("[table-wrap] th").eq(stickyL).attr('stickyL', '');
        }
    }
    if (!Number.isNaN(stickyR)) {
        $("[table-wrap] th[stickyR]").removeAttr('stickyR');
        if (stickyR != -1) {
            $("[table-wrap] th").eq(stickyR).attr('stickyR', '');
        }
    }
    tableWrapHeighter();
}

// 计算剩余视口高度并窗口尺寸变化时更新
function calcInnerRestHeight(restHeight, heightSetter) {
    if (typeof heightSetter !== 'function') {
        return
    }
    function innerHeighter() {
        var height = window.innerHeight - restHeight;
        heightSetter(height);
    }
    
    const innerHeighterDebounce = debounce(innerHeighter, 500);
    window.addEventListener('resize', innerHeighterDebounce);
    innerHeighter();
}

// 发送 POST 请求
function fetchPOST(url, body) {
    return new Promise((resolve, reject) => {
        fetch(url, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(body || {}) // 请求体
        })
            .then(r => {
                if (!r.ok) {
                    throw new Error(`${r.status} ${r.statusText}`);
                }
                return r.json();
            })
            .then(res => {
                if (res.code !== 200) {
                    return reject(res);
                }
                resolve(res);
            })
            .catch(error => {
                // console.error('fetchPOST Error:', error); // 处理错误
                reject({ msg: error });
            });
    })
}

// 发送 GET 请求
function fetchGET(url, body) {
    return new Promise((resolve, reject) => {
        fetch(url)
            .then(r => {
                if (!r.ok) {
                    throw new Error(`${r.status} ${r.statusText}`);
                }
                return r.json();
            })
            .then(res => {
                if (res.code !== 200) {
                    return reject(res);
                }
                resolve(res);
            })
            .catch(error => {
                // console.error('fetchGET Error:', error); // 处理错误
                reject({ msg: error });
            });
    })
}

// 下载文件
function fetchDownloadFile(url, body) {
    return new Promise((resolve, reject) => {
        var fileName = '';
        var f;
        if (!body) {
            f = fetch(url)
        } else {
            f = fetch(url, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify(body || {}) // 请求体
            })
        }
        f.then(r => {
            if (!r.ok) {
                if (r.status == 416) {
                    throw new Error('未查询到可以导出的数据'); 
                }
                throw new Error(`${r.status} ${r.statusText}`); 
            }

            var disposition = r.headers.get('Content-Disposition');
            if (disposition) {
                fileName = decodeURIComponent(disposition.replace('attachment;filename=', ''));
            }
            return r.blob();
        })
        .then(blob => {
            // 创建一个指向 Blob 的 URL
            var blobUrl = window.URL.createObjectURL(blob);

            // 创建一个 <a> 标签
            var link = document.createElement('a');
            link.href = blobUrl;
            link.download = fileName; // 设置下载的文件名

            // 触发点击事件以下载文件
            document.body.appendChild(link);
            link.click();

            // 清理 URL 对象
            window.URL.revokeObjectURL(blobUrl);
            document.body.removeChild(link);
            resolve({fileName: fileName});
        })
        .catch(error => {
            // console.error('fetchDownloadFile Error:', error); // 处理错误
            reject({ msg: error });
        });
    })
}

// 发送 FormData 请求
function fetchFormData(url, formData) {
    return new Promise((resolve, reject) => {
        fetch(url, {
            method: 'POST',
            body: formData // 请求体
        })
            .then(r => {
                if (!r.ok) {
                    throw new Error(`${r.status} ${r.statusText}`);
                }
                return r.json();
            })
            .then(res => {
                if (res.code !== 200) {
                    return reject(res);
                }
                resolve(res);
            })
            .catch(error => {
                // console.error('fetchPOST Error:', error); // 处理错误
                reject({ msg: error });
            });
    })
}