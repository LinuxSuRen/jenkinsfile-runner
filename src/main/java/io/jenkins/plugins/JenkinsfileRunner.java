package io.jenkins.plugins;

import hudson.Extension;
import hudson.init.InitMilestone;
import hudson.init.Initializer;
import hudson.model.Result;
import hudson.model.TaskListener;
import hudson.model.TopLevelItem;
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

/**
 * @author <a href="mailto:nicolas.deloof@gmail.com">Nicolas De Loof</a>
 */
@Extension
public class JenkinsfileRunner extends RunListener<WorkflowRun> {


    @Initializer(after = InitMilestone.JOB_LOADED, displayName = "jenkinsfile-runner", fatal = true)
    public static void run() {

        try {
            Jenkins j = Jenkins.getInstance();
            final TopLevelItem job = j.getItem("job");
            if (job != null) job.delete();

            WorkflowJob w = j.createProject(WorkflowJob.class, "job");
            w.addProperty(new DurabilityHintJobProperty(FlowDurabilityHint.PERFORMANCE_OPTIMIZED));
            final File jenkinsfile = new File("./Jenkinsfile");
            w.setDefinition(new CpsScmFlowDefinition(
                    new FileSystemSCM(jenkinsfile.getParentFile()), jenkinsfile.getName()));
            QueueTaskFuture<WorkflowRun> f = w.scheduleBuild2(0,
                    new SetJenkinsfileLocation(jenkinsfile));
        } catch (Throwable t) {
            t.printStackTrace();
            System.exit(2);
        }
    }

    @Override
    public void onStarted(WorkflowRun run, TaskListener listener) {
        try {
            run.writeWholeLogTo(System.out);
        } catch (IOException | InterruptedException e) {
            System.err.println("Failed to redirect build log to stdout: " + e.getMessage());
            System.exit(127);
        }
    }

    @Override
    public void onCompleted(WorkflowRun run, @Nonnull TaskListener listener) {
        System.exit(run.getResult().isBetterOrEqualTo(Result.SUCCESS) ? 0 : -1);
    }

}
