require(dplyr)
require(ggplot2)
require(quantileCI)
require(base64enc)

read.al <- function(path) {
  df <- read.csv(path, sep=",",header=T, dec=".")
  return(df) #(tail(df, -500))
}

summary_table <- function(df1, tag1, df2, tag2) {
  qCI <- function(df, p) {
    return(quantileCI::quantile_confint_nyblom(df, p))
  }
  stats <- function(df) {
    avg = signif(t.test(df)$conf.int, digits = 2)
    p50 = signif(qCI(df, 0.5), digits = 4)
    p95 = signif(qCI(df, 0.95), digits = 4)
    p99 = signif(qCI(df, 0.99), digits = 4)
    p999 = signif(qCI(df, 0.999), digits = 4)
    p9999 = signif(qCI(df, 0.9999), digits = 4)
    p99999 = signif(qCI(df, 0.99999), digits = 4)
    dist = signif(qCI(df, 0.99999)- qCI(df, 0.5), digits = 4)
    data <- c(avg, p50, p95, p99, p999, p9999, p99999, dist)
    return(data)
  }
  
  stats1 = stats(df1)
  stats2 = stats(df2)
  avgdf    <- data.frame("avg",    stats1[1],  stats1[2],  stats2[1],  stats2[2])
  p50df    <- data.frame("p50",    stats1[3],  stats1[4],  stats2[3],  stats2[4])
  p95df    <- data.frame("p95",    stats1[5],  stats1[6],  stats2[5],  stats2[6])
  p99df    <- data.frame("p99",    stats1[7],  stats1[8],  stats2[7],  stats2[8])
  p999df   <- data.frame("p999",   stats1[9],  stats1[10], stats2[9],  stats2[10])
  p9999df  <- data.frame("p9999",  stats1[11], stats1[12], stats2[11], stats2[12])
  p99999df <- data.frame("p99999", stats1[13], stats1[14], stats2[13], stats2[14])
  distdf   <- data.frame("dist",   stats1[15], stats1[16], stats2[15], stats2[16])
  
  tag1_inf = paste(tag1, "cii", sep = ".")
  tag1_sup = paste(tag1, "cis", sep = ".")
  tag2_inf = paste(tag2, "cii", sep = ".")
  tag2_sup = paste(tag2, "cis", sep = ".")
  names(avgdf)    <- c("stats", tag1_inf, tag1_sup, tag2_inf, tag2_sup)
  names(p50df)    <- c("stats", tag1_inf, tag1_sup, tag2_inf, tag2_sup)
  names(p95df)    <- c("stats", tag1_inf, tag1_sup, tag2_inf, tag2_sup)
  names(p99df)    <- c("stats", tag1_inf, tag1_sup, tag2_inf, tag2_sup)
  names(p999df)   <- c("stats", tag1_inf, tag1_sup, tag2_inf, tag2_sup)
  names(p9999df)  <- c("stats", tag1_inf, tag1_sup, tag2_inf, tag2_sup)
  names(p99999df) <- c("stats", tag1_inf, tag1_sup, tag2_inf, tag2_sup)
  names(distdf)   <- c("stats", tag1_inf, tag1_sup, tag2_inf, tag2_sup)
  df <- rbind(avgdf, p50df, p95df, p99df, p999df, p9999df, p99999df, distdf)
  df
}

graph_tail <- function(gci, nogci, title, x_limit_sup) {
  cmp <- rbind(
    data.frame("response_time"=gci, Type="GCI"),
    data.frame("response_time"=nogci, Type="NOGCI")
  )
  gci.color <- "blue"
  gci.p999 <- quantile(gci, 0.9999)
  gci.p50 <- quantile(gci, 0.5)
  
  nogci.color <- "red"
  nogci.p999 <- quantile(nogci, 0.9999)
  nogci.p50 <- quantile(nogci, 0.5)
  
  annotate_y = 0.9
  size = 0.5
  alpha = 0.5
  angle = 90
  p <- ggplot(cmp, aes(response_time, color=Type)) +
    stat_ecdf(size=size) +
    # P50
    annotate(geom="text", x=gci.p50, y=annotate_y, label="Median", angle=angle, color=gci.color) +
    geom_vline(xintercept=gci.p50, linetype="dotted", size=size, alpha=alpha, color=gci.color) +
    annotate(geom="text", x=nogci.p50, y=annotate_y, label="Median", angle=angle, color=nogci.color) + 
    geom_vline(xintercept=nogci.p50, linetype="dotted", size=size, alpha=alpha, color=nogci.color) +
    
    # P999
    annotate(geom="text", x=gci.p999, y=annotate_y, label="99.99th", angle=angle, color=gci.color) +
    geom_vline(xintercept=gci.p999, linetype="dotted", size=size, alpha=alpha, color=gci.color) +
    annotate(geom="text", x=nogci.p999, y=annotate_y, label="99.99th", angle=angle, color=nogci.color) + 
    geom_vline(xintercept=nogci.p999, linetype="dotted", size=size, alpha=alpha, color=nogci.color) +
    
    #scale_x_continuous(breaks=seq(0, max(cmp$latency), 10)) +
    #coord_cartesian(ylim = c(0.99, 1)) +
    xlim(0, x_limit_sup) +
    theme(legend.position="top") +
    scale_color_manual(breaks = c("GCI", "NOGCI"), values=c("blue", "red")) +
    theme_bw() +
    ggtitle(title) +
    xlab("response time (ms)") +
    ylab("rate") 
  
  print(p)
}
