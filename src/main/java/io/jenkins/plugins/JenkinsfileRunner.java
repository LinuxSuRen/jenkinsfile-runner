package io.jenkins.plugins;

import hudson.Extension;
import hudson.init.InitMilestone;
import hudson.init.Initializer;
import hudson.model.Result;
import hudson.model.Run;
import hudson.model.TaskListener;
import hudson.model.listeners.RunListener;
import hudson.model.queue.QueueTaskFuture;
import jenkins.model.Jenkins;
import org.jenkinsci.plugins.workflow.cps.CpsScmFlowDefinition;
import org.jenkinsci.plugins.workflow.flow.FlowDurabilityHint;
import org.jenkinsci.plugins.workflow.job.WorkflowJob;
import org.jenkinsci.plugins.workflow.job.WorkflowRun;
import org.jenkinsci.plugins.workflow.job.properties.DurabilityHintJobProperty;

import javax.annotation.Nonnull;
import java.io.File;
import java.io.IOException;
import java.util.concurrent.ExecutionException;

/**
 * @author <a href="mailto:nicolas.deloof@gmail.com">Nicolas De Loof</a>
 */
public class JenkinsfileRunner {


    @Initializer(after = InitMilestone.JOB_LOADED, displayName = "jenkinsfile-runner", fatal = true)
    public static void run() throws IOException, ExecutionException, InterruptedException {
        Jenkins j = Jenkins.getInstance();
        WorkflowJob w = j.createProject(WorkflowJob.class, "job");
        w.addProperty(new DurabilityHintJobProperty(FlowDurabilityHint.PERFORMANCE_OPTIMIZED));
        final File jenkinsfile = new File("./Jenkinsfile");
        w.setDefinition(new CpsScmFlowDefinition(
                new FileSystemSCM(jenkinsfile.getParentFile()), jenkinsfile.getName()));
        QueueTaskFuture<WorkflowRun> f = w.scheduleBuild2(0,
                new SetJenkinsfileLocation(jenkinsfile));
    }

    @Extension
    public final static RunListener LISTENER = new RunListener() {

        @Override
        public void onStarted(Run run, TaskListener listener) {
            try {
                run.writeWholeLogTo(System.out);
            } catch (IOException | InterruptedException e) {
                System.err.println("Failed to redirect build log to stdout: " + e.getMessage());
                System.exit(127);
            }
        }

        @Override
        public void onCompleted(Run run, @Nonnull TaskListener listener) {
            System.exit(run.getResult().isBetterOrEqualTo(Result.SUCCESS) ? 0 : -1);
        }
    };

}
