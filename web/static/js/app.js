$(function () {
    if ($("#numbers-section").length) {
        loadDashboard("2");
        $("#day-selector").change(function(){
            loadDashboard($(this).val());
        });
    };

    if ($("#list-selector").length) {
        $("#list-selector").change(function(){
            loadDay($(this).val(), 0); // on change
        });
        $("#day-list-prev, #day-list-next").click(function(e){
            e.preventDefault();
            loadDay($("#list-selector").val(), $(this).data("page")); // on click
        });

        loadDay($("#list-selector").val(), 0); // on load
    };
});

function loadDay(listType, page) {
    $(".after-load").hide();
    var selectedDate = $("#selectedDate").val();
    var table = $("#events-table tbody");
    table.empty();
    queryURL = "/data/day/" + selectedDate + "/list/" + listType + "/page/" + page;
    // console.log("Query URL: " + queryURL);
    $.get(queryURL, function (data) {
        //console.log(data);

        var prevLink = $("#day-list-prev");
        var nextLink = $("#day-list-next");

        prevLink.data("page", data.pagePrev);
        nextLink.data("page", data.pageNext);

        if (data.hasPrev) {
            prevLink.show();
        }else{
            prevLink.hide();
        }

        if (data.hasNext) {
            nextLink.show();
        }else{
            nextLink.hide();
        }
        
        $.each(data.events, function(rowIndex, e) {
            var row = $(`<tr class="user-data-row" data-user="${e.username}"/>`);
            row.append(`<td class="user-img">
                    <a href="#" class="no-link" 
                       title="${e.description} - (updated: ${e.updated_at})">
                        <img src="${e.profile_image}" class="profile-image" />
                    </a>
                </td>`);
            row.append(`<td class="user-name">
                <a href="#" class="no-link" 
                   title="${e.description} - (updated: ${e.updated_at})">
                    @${e.username}</a><div>${e.name}<br />${e.location}</div>
                </td>`);
            row.append(`<td class="user-data"><div>${e.has_relation}</div></td>`); 
            row.append(`<td class="user-data"><div>${e.friend_count}</div></td>`); 
            row.append(`<td class="user-data"><div>${e.followers_count}</div></td>`); 
            row.append(`<td class="user-data"><div>${e.post_count}</div></td>`); 
            row.append(`<td class="user-data"><div>${e.listed_count}</div></td>`); 
            table.append(row);
        });

        $(".no-link").click(function(e){
            e.preventDefault();
        });

        $(".user-data-row").on("mouseover", function() {
            $(this).closest("tr").addClass("highlight");
            $(this).closest("table").find(".user-data-row:nth-child(" + ($(this).index() + 1) + ")").addClass("highlight");
        });
    
        $(".user-data-row").on("mouseout", function() {
            $(this).closest("tr").removeClass("highlight");
            $(this).closest("table").find(".user-data-row:nth-child(" + ($(this).index() + 1) + ")").removeClass("highlight");
        });

        $(".user-data-row").click(function(){
            window.open("https://twitter.com/" + $(this).data('user'), "_blank");
        });
    
    }).fail(function(jqXHR, textStatus, errorThrown) {
        console.log(jqXHR);
        console.log(textStatus);
        console.log(errorThrown);
        $("#error-msg").html("Error loading date data, see logs for details.").show()
    });
}

function loadDashboard(days) {
    $(".after-load").hide();
    // console.log("period days: " + days);
    $.get("/data/dash?days=" + days, function (data) {
        // console.log(data);

        // numbers
        $("#follower-count .data").text(data.state.follower_count).digits();
        $("#friend-count .data").text(data.state.friend_count).digits();
        $("#follower-gained-count .data").text(data.state.new_follower_count).digits();
        $("#follower-lost-count .data").text(data.state.new_unfollower_count).digits();
        $("#listed-count .data").text(data.user.listed_count).digits();
        $("#post-count .data").text(data.user.post_count).digits();
        $("#meta-updated-on").text(data.updated_on);

        $(".wait-load").hide();
        $(".after-load").show();

        // follower count chart
        $("#follower-event-series").remove();
        $("#follower-chart").append('<canvas id="follower-event-series"></canvas>');
        var followerChart = new Chart($("#follower-event-series")[0].getContext("2d"), {
            type: 'bar',
            data: {
                labels: Object.keys(data.series.new_followers),
                datasets: [{
                    label: 'Unfollowed',
                    data: Object.values(data.series.lost_followers),
                    backgroundColor: 'rgba(206, 149, 166,0.1)',
                    borderColor: 'rgba(206, 149, 166,0.5)',
                    borderWidth: 1,
                    minBarLength: 2,
                },
                {
                    label: 'Followed',
                    data: Object.values(data.series.new_followers),
                    backgroundColor: 'rgba(127, 201, 143,0.1)',
                    borderColor: 'rgba(127, 201, 143,0.5)',
                    borderWidth: 1,
                    minBarLength: 2,
                },
                {
                    label: 'Friended',
                    data: Object.values(data.series.new_friends),
                    backgroundColor: 'rgba(127, 201, 143,0.4)',
                    borderColor: 'rgba(127, 201, 143,0.7)',
                    borderWidth: 1,
                    minBarLength: 2
                }, {
                    label: 'Unfriended',
                    data: Object.values(data.series.lost_friends),
                    backgroundColor: 'rgba(206, 149, 166,0.4)',
                    borderColor: 'rgba(206, 149, 166,0.7)',
                    borderWidth: 1,
                    minBarLength: 2
                }, 
                {
                    label: 'Follower Average',
                    type: 'line',
                    fill: false,
                    data: Object.values(data.series.avg_followers),
                    backgroundColor: 'rgba(255, 255, 204,0.4)',
                    borderColor: 'rgba(255, 255, 204,0.4)',
                    borderWidth: 2,
                }
                ]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                title: {
                    display: true,
                    text: 'Events - whom you (un)followed and who (un)friended you',
                    fontColor: 'rgba(250, 250, 250, 0.5)',
                    fontSize: 16,
                },
                legend: {
                    display: true,
                    position: 'bottom',
                    labels: {
                        fontSize: 16
                    }
                },
                scales: {
                    yAxes: [
                        {
                            ticks: {
                                beginAtZero: false,
                                fontColor: 'rgba(250, 250, 250, 0.5)',
                                fontSize: 14,
                                maxTicksLimit: 7,
                                precision: 0
                            },
                            stacked: true
                        }
                    ],
                    xAxes: [
                        {
                            ticks: {
                                beginAtZero: false,
                                fontColor: 'rgba(250, 250, 250, 0.5)',
                                fontSize: 14
                            },
                            stacked: true
                        }
                    ]
                },
                onClick: (evt, item) => {
                    if (item.length) {
                        var model = item[0]._model;
                        console.log("Date: ", model.label);
                        redirectToDate(model.label);
                    }
                }
            }
        });

        // follower count chart
        $("#follower-count-series").remove();
        $("#follower-count-chart").append('<canvas id="follower-count-series"></canvas>');
        var followerChart = new Chart($("#follower-count-series")[0].getContext("2d"), {
            type: 'line',
            data: {
                labels: Object.keys(data.series.all_followers),
                datasets: [{
                    label: 'Count',
                    data: Object.values(data.series.all_followers),
                    backgroundColor: 'rgba(109, 110, 110, 0.4)',
                    borderColor: 'rgba(109, 110, 110, 1)',
                    minBarLength: 2,
                    borderWidth: 1
                }, 
                {
                    label: 'Average',
                    fill: false,
                    data: Object.values(data.series.avg_total),
                    backgroundColor: 'rgba(255, 255, 204,0.4)',
                    borderColor: 'rgba(255, 255, 204,0.4)',
                    borderWidth: 2,
                }]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                title: {
                    display: true,
                    text: 'Totals - number of followers per day',
                    fontColor: 'rgba(250, 250, 250, 0.5)',
                    fontSize: 16,
                },
                legend: {
                    display: false
                },
                scales: {
                    yAxes: [
                        {
                            ticks: {
                                beginAtZero: false,
                                fontColor: 'rgba(250, 250, 250, 0.5)',
                                fontSize: 14,
                                maxTicksLimit: 7,
                                precision: 0
                            }
                        }
                    ],
                    xAxes: [
                        {
                            ticks: {
                                beginAtZero: false,
                                fontColor: 'rgba(250, 250, 250, 0.5)',
                                fontSize: 14
                            }
                        }
                    ]
                }
            }
        });




    }).fail(function(jqXHR, textStatus, errorThrown) {
        console.log(jqXHR);
        console.log(textStatus);
        console.log(errorThrown);
        console.log("error loading twitter data");
        // $(location).attr("href", "/logout");
    });
}

function redirectToDate(d) {
    $(location).attr("href", "/view/day/" + d);
}

function makeLinks() {
    var tweetText = $(".tweet-text");
    if (tweetText.length) {
        tweetText.each(
            function () {
                var $words = $(this).text().split(' ');
                for (i in $words) {
                    if ($words[i].indexOf('https://t.co/') == 0) {
                        $words[i] = '<a href="' + $words[i] + '" target="_blank">' + $words[i] + '</a>';
                    }
                }
                $(this).html($words.join(' '));
            }
        );
    }
}

$.fn.digits = function () {
    return this.each(function () {
        $(this).text($(this).text().replace(/(\d)(?=(\d\d\d)+(?!\d))/g, "$1,"));
    });
}